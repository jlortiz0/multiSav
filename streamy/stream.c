#include "stream.h"

#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/frame.h>
#include <libavutil/imgutils.h>
#include <libswscale/swscale.h>
#include <stdlib.h>

// TODO: Allow custom headers
typedef struct LibavReader {
    AVFormatContext *context;
    AVCodecContext *codec;
    AVFrame *frame;
    AVPacket *packet;
    struct SwsContext *scaler;
    AVFrame *frame2;
    int idx;
} LibavReader;

int libavreader_new(const char *fName, LibavReader **ptr, char *user_agent) {
    LibavReader *thing = malloc(sizeof(LibavReader));
    thing->context = avformat_alloc_context();
    AVDictionary *dict = NULL;
    if (user_agent != NULL) {
        av_dict_set(&dict, "user_agent", user_agent, 0);
    }
    int code = avformat_open_input(&thing->context, fName, NULL, &dict);
    if (dict != NULL && av_dict_count(dict) != 0) {
        abort();
    }
    if (code) {
        avformat_free_context(thing->context);
        free(thing);
        return code;
    }
    code = avformat_find_stream_info(thing->context, NULL);
    if (code) {
        avformat_free_context(thing->context);
        free(thing);
        return code;
    }
    thing->codec = NULL;
    for (unsigned int i = 0; i < thing->context->nb_streams; i++) {
        if (thing->context->streams[i]->codecpar->codec_type == AVMEDIA_TYPE_VIDEO) {
            const AVCodec *c = avcodec_find_decoder(thing->context->streams[i]->codecpar->codec_id);
            if (!c) {
                return -1;
            }
            thing->codec = avcodec_alloc_context3(NULL);
            avcodec_parameters_to_context(thing->codec, thing->context->streams[i]->codecpar);
            code = avcodec_open2(thing->codec, c, NULL);
            if (code) {
                avcodec_free_context(&thing->codec);
                avformat_free_context(thing->context);
                free(thing);
                return code;
            }
            thing->idx = i;
            break;
        }
    }
    if (thing->codec == NULL) {
        avformat_free_context(thing->context);
        free(thing);
        return -1;
    }
    thing->frame = av_frame_alloc();
    thing->packet = av_packet_alloc();
    if (thing->codec->pix_fmt != AV_PIX_FMT_RGBA) {
        thing->scaler = sws_getContext(thing->codec->width, thing->codec->height, thing->codec->pix_fmt,
                                       thing->codec->width, thing->codec->height, AV_PIX_FMT_RGBA, 0, NULL, NULL, 0);
        thing->frame2 = av_frame_alloc();
        thing->frame2->height = thing->codec->height;
        thing->frame2->width = thing->codec->width;
        thing->frame2->format = AV_PIX_FMT_RGBA;
    } else {
        thing->scaler = NULL;
    }
    *ptr = thing;
    return 0;
}

int libavreader_next(LibavReader *l, uint8_t *buf) {
    int code = avcodec_receive_frame(l->codec, l->frame);
    if (code) {
        if (code == AVERROR(EAGAIN)) {
            code = av_read_frame(l->context, l->packet);
            if (code == AVERROR_EOF) {
                avio_seek(l->context->pb, 0, SEEK_SET);
                code = avformat_seek_file(l->context, l->idx, 0, 0, 0, 0);
                if (code) {
                    return code;
                }
                return libavreader_next(l, buf);
            } else if (code) {
                return code;
            }
            if (l->packet->stream_index != l->idx) {
                av_packet_unref(l->packet);
                return libavreader_next(l, buf);
            }
            code = avcodec_send_packet(l->codec, l->packet);
            av_packet_unref(l->packet);
            if (code) {
                return code;
            }
            return libavreader_next(l, buf);
        }
        return code;
    }
    AVFrame *toCopy = l->frame;
    if (l->scaler != NULL && buf != NULL) {
        l->frame2->height = l->codec->height;
        l->frame2->width = l->codec->width;
        l->frame2->format = AV_PIX_FMT_RGBA;
        code = av_frame_get_buffer(l->frame2, 0);
        code = sws_scale(l->scaler, (const uint8_t *const *)l->frame->data, l->frame->linesize, 0,
                         l->frame->height, (uint8_t *const *)l->frame2->data, l->frame2->linesize);
        // code = sws_scale_frame(l->scaler, l->frame2, l->frame);
        if (code < 0) {
            return code;
        }
        toCopy = l->frame2;
        av_frame_unref(l->frame);
    }
    if (buf != NULL) {
        av_image_copy_to_buffer(buf, toCopy->height * toCopy->width * 4,
                                (const uint8_t *const *)toCopy->data, toCopy->linesize, toCopy->format, toCopy->width,
                                toCopy->height, 4);
    }
    av_frame_unref(toCopy);
    return 0;
}

pair_int libavreader_dimensions(const LibavReader *l) {
    pair_int p = {0};
    if (l == NULL || l->codec == NULL) {
        return p;
    }
    p.x = l->codec->width;
    p.y = l->codec->height;
    return p;
}

void libavreader_destroy(LibavReader *l) {
    if (l != NULL) {
        if (l->scaler != NULL) {
            av_frame_free(&l->frame2);
            sws_freeContext(l->scaler);
        }
        av_frame_free(&l->frame);
        av_packet_free(&l->packet);
        avcodec_close(l->codec);
        avcodec_free_context(&l->codec);
        avformat_close_input(&l->context);
        free(l);
    }
}

float libavreader_fps(const LibavReader *l) {
    int fr = av_q2intfloat(l->context->streams[l->idx]->avg_frame_rate);
    return *((float *)&fr);
}
