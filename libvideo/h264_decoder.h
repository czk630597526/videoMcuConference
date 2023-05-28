/*
 * h264_decoder.h
 *
 */

#ifndef H264_DECODER_C_
#define H264_DECODER_C_
#include <stdint.h>
#include "libv_log.h"
#include "libvideo.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef enum __H264_DECODER_ID
{
    H264_DECODER_ID_FFM = 0,    //  ffmpeg
    H264_DECODER_ID_OPENH264 = 1,    //openh264
    H264_DECODER_ID_MAX = 2,
}H264_DECODER_ID;

typedef struct __H264_DECODER_TH    H264_DECODER_TH;
struct __H264_DECODER_TH
{
    H264_DECODER_ID _id;     //解码器的标识，上层不用处理。
    void *_private_data;     //   每个解码器的私有数据结构体，上层不能修改。

    //解码处理函数：
    /*
    pdecth：     解码器的句柄；
    pout_pic：输出数据，其中的data上层不用申请，也不用释放，decoder会负责维护。 解码后data在下一次调用decode_video函数
                                     前都是可用的，如果需要保存下来，则需要调用libvideo_copy_frame_pic进行拷贝。这样做的目的是为了减少内存拷贝。
                                     输出的时戳取决于输入的时戳。因为输出的帧往往不是输入的帧。
    got_picture_ptr:释放有帧输出，为0表示此次没有帧输出，1表示有帧输出。
    pnal    ：输入数据，如果携带时戳，则会在这帧输出时将时戳赋值。
    return  : 失败为-1，大于0表示解码的byte数，0表示没有解码。
     */
    int32_t (*decode_video)(H264_DECODER_TH *pdecth, FRAME_PIC_DATA *pout_pic, int32_t *got_picture_ptr, const H264_NAL_DATA *pnal);

};

//工厂方法结构体，参考设计模式：工厂方法.?有些不确定模式名称是否叫工厂
typedef struct __H264_DECODER_FACTORY
{
    H264_DECODER_ID id;     //解码器的标识
    int32_t (*init)(H264_DECODER_TH *pdecth);      //初始化函数，pdecth对象由上层申请。
    void (*release)(H264_DECODER_TH *pdecth);   //释放函数，类似于析构函数，pdecth对象由上层释放。

}H264_DECODER_FACTORY;

void h264_decoder_init();   //上层不用调用，会在libvideo_init中调用。
int32_t h264_decoder_register(H264_DECODER_FACTORY *pdec_fact);//内部使用，上层不用调用

//根据解码器id申请一个解码器，需要调用free_decoder是否。
H264_DECODER_TH *alloc_decoder_by_id(H264_DECODER_ID id);
void free_decoder(H264_DECODER_TH *pdecth);


#ifdef __cplusplus
}
#endif

#endif /* H264_DECODER_C_ */
