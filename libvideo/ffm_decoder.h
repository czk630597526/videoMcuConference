/*
 * ffm_decoder.h
 *
 */

#ifndef FFM_DECODER_H_
#define FFM_DECODER_H_

#include <stdint.h>
#include "libv_log.h"
#include "libvideo.h"

#ifdef __cplusplus
extern "C" {
#endif


void ffm_decoder_init();    //上层不用调用，会在libvideo_init中调用。

#ifdef __cplusplus
}
#endif


#endif /* FFM_DECODER_H_ */
