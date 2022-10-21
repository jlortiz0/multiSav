#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/frame.h>
#include <libswscale/swscale.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>

typedef struct LibavReader {
    AVFormatContext *context;
    AVCodecContext *codec;
    AVFrame *frame;
    AVPacket *packet;
    struct SwsContext *scaler;
    int idx;
} LibavReader;

int libavreader_new(char *fName, LibavReader **ptr) {
    LibavReader *thing = malloc(sizeof(LibavReader));
    thing->context = avformat_alloc_context();
    int code = avformat_open_input(&thing->context, fName, NULL, NULL);
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
    for (int i = 0; i < thing->context->nb_streams; i++) {
        if (thing->context->streams[i]->codecpar->codec_type == AVMEDIA_TYPE_VIDEO) {
            AVCodec *c = avcodec_find_decoder(thing->context->streams[i]->codecpar->codec_id);
            if (!c) {
                return -1;
            }
            thing->codec = avcodec_alloc_context3(NULL);
            avcodec_parameters_to_context(thing->codec, thing->context->streams[i]->codecpar);
            code = avcodec_open2(thing->codec, c, NULL);
            if (code) {
                avcodec_free_context(thing->codec);
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
        thing->scaler = sws_getContext(thing->codec->width, thing->codec->height, thing->codec->pix_fmt, thing->codec->width, thing->codec->height, AV_PIX_FMT_RGBA, 0, NULL, NULL, 0);
    }
    *ptr = thing;
    return 0;
}

int libavreader_next(LibavReader *l, uint8_t *buf) {
    int code = avcodec_receive_frame(l->codec, l->frame);
    if (code) {
        if (code == AVERROR(EAGAIN)) {
            code = av_read_frame(l->context, l->packet);
            if (code) {
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
            return libavreader_new(l, buf);
        }
        return code;
    }
    if (l->scaler != NULL) {
        sws_scale(l->scaler, (const uint8_t *const *)l->frame->buf, l->frame->linesize, 0, l->frame->height, buf, 0);
    } else {
        memcpy(buf, (const uint8_t *)l->frame->buf, 4 * l->frame->height * l->frame->width);
    }
    return 0;
}

typedef struct {
    int x;
    int y;
} pair_int;

pair_int libavreader_dimensions(LibavReader *l) {
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
        sws_freeContext(l->scaler);
        av_frame_free(&l->frame);
        av_packet_free(&l->packet);
        avcodec_close(l->codec);
        avcodec_free_context(&l->context);
        avformat_close_input(&l->context);
    }
}
