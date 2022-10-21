ifeq ($(OS),Windows_NT)
CC=gcc
else
CC=clang
endif
CFLAGS=$(shell pkg-config --cflags libavformat libavcodec libavutil libswscale) -Wall -Wextra -Wpedantic -Werror -g
LDFLAGS=$(shell pkg-config --libs libavformat libavcodec libavutil libswscale)

.PHONY: all clean format

all: stream.o

clean:
	rm *.o test test.exe out.tga -f

%.o: %.c
	$(CC) $(CFLAGS) $< -c

test: stream.o test.o
	$(CC) $(LDFLAGS) $^ -o test

format:
	clang-format -i *.c *.h

scan-build: clean
	scan-build --use-cc=clang make test
