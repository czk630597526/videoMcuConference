/*
 * libvideo.h
 *
 */

#ifndef LIBVIDEO_H_
#define LIBVIDEO_H_

#include "libv_log.h"
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif


typedef enum __NALU_TYPE
{
    NALU_TYPE_UNDEFINE = 0,                    // 0：未规定
    NALU_TYPE_SLICE    = 1 ,                   // 1：非IDR图像中不采用数据划分的片段
    NALU_TYPE_DPA      = 2 ,                   // 2：非IDR图像中A类数据划分片段
    NALU_TYPE_DPB      = 3 ,                   // 3：非IDR图像中B类数据划分片段
    NALU_TYPE_DPC      = 4 ,                   // 4：非IDR图像中C类数据划分片段
    NALU_TYPE_IDR      = 5 ,                   // 5：IDR图像的片段
    NALU_TYPE_SEI      = 6 ,                   // 6：补充增强信息 (SEI)
    NALU_TYPE_SPS      = 7 ,                   // 7：序列参数集
    NALU_TYPE_PPS      = 8 ,                   // 8：图像参数集
    NALU_TYPE_AUD      = 9 ,                   // 9：分割符
    NALU_TYPE_EOSEQ    = 10,                   // 10：序列结束符
    NALU_TYPE_EOSTREAM = 11,                   // 11：流结束符
    NALU_TYPE_FILL     = 12,                   // 12：填充数据
                                               //13 – 23：保留
                                               //24 – 31：应用于网络分包
}NALU_TYPE;

typedef enum __FRAME_TYPE
{
     FRAME_TYPE_INVALID,
     FRAME_TYPE_IDR,
     FRAME_TYPE_I,
     FRAME_TYPE_P,
     FRAME_TYPE_SKIP,
     FRAME_TYPE_I_P_MIXED,
     FRAME_TYPE_B,
}FRAME_TYPE;

typedef enum __VIDEO_MIX_TYPE
{
    VIDEO_MIX_FULL = 1,
    VIDEO_MIX_TWO  = 2,
    VIDEO_MIX_THREE  = 3,
    VIDEO_MIX_FOUR = 4,
    VIDEO_MIX_SIX = 6,
    VIDEO_MIX_EIGHT = 8,
    VIDEO_MIX_NINE = 9,
    VIDEO_MIX_THIRTEEN = 13,
    VIDEO_MIX_SIXTEEN = 16,
    VIDEO_MIX_TWENTY_FIVE = 25,
    VIDEO_MIX_MAX = 26,
}VIDEO_MIX_TYPE;

typedef struct __H264_NAL_PACKET
{
    uint32_t used_flag;         //是否使用：0：未使用，1：使用。上层程序不用使用此字段。
    NALU_TYPE nal_type;          //nal类型：NALU_TYPE枚举
    FRAME_TYPE frame_type;       //帧的类型，和nal的类型是不同的，指明是I,P,B帧等
    uint32_t  timestamp;         //nal的时戳
    uint32_t data_size;         //nal包的大小
    uint8_t *data_buf;        //nal包指针，指向一块临时缓冲区，不支持多线程，如果需要存储则要再申请内存。
                                    //格式：[0,0,0,1]Start sequence + [nal header, data]NAL data
    uint32_t frame_end_flg;      //帧结束标志位：用于一个nal一个slice，发包是单个slice模式，用于标识帧结束。
                                 //编码时输出。上层不用设置；
}H264_NAL_DATA;

typedef struct __H264_RTP_PACKET
{
    uint32_t used_flag;         //是否使用：0：未使用，1：使用。上层程序不用使用此字段。
    uint32_t seq;               //rtp包的序号
    uint32_t data_size;         //rtp包的大小
    uint8_t *data_buf;          //rtp包指针，指向一块临时缓冲区，不支持多线程，如果需要存储则要再申请内存。
}H264_RTP_PACKET;

typedef enum __PICTURE_TYPE
{
    PICTURE_TYPE_NONE = 0, ///< Undefined
    PICTURE_TYPE_I,     ///< Intra
    PICTURE_TYPE_P,     ///< Predicted
    PICTURE_TYPE_B,     ///< Bi-dir predicted
    PICTURE_TYPE_S,     ///< S(GMC)-VOP MPEG4
    PICTURE_TYPE_SI,    ///< Switching Intra
    PICTURE_TYPE_SP,    ///< Switching Predicted
    PICTURE_TYPE_BI,    ///< BI type
}PICTURE_TYPE;

typedef enum __PICTURE_FORMAT
{
    PIC_FMT_NONE = -1,
    PIC_FMT_YUV_I420,   ///< planar YUV 4:2:0, 12bpp, (1 Cr & Cb sample per 2x2 Y samples)
}PICTURE_FORMAT;

#define LIBV_NUM_DATA_POINTERS    8

//保存原始图像数据，比如yuv,解码后的输出
//可以使用函数libvideo_alloc_frame_pic申请，但是需要调用libvideo_free_frame_pic进行释放。
//也可以自己直接构造这个结构体，data和linesize用解码后的输出。
typedef struct __FRAME_PIC_DATA
{
    PICTURE_FORMAT pic_format;      //表示存储格式，目前只支持420p
    PICTURE_TYPE pic_type;          //图片帧格式，是：i，p，b或其他。解码的输出。
    uint32_t  timestamp;         //图片的时戳
    int32_t   key_frame;            //释放时关键帧
    int32_t   quality;              //图片质量
    int32_t   width;
    int32_t   height;
    int32_t _alloc_flg;             //data中的内存是否是库中申请的。释放时，如果此标志为1，则释放data。上层不要使用
    uint8_t *data[LIBV_NUM_DATA_POINTERS];      //可以使用去申请内存,可以用av_image_alloc申请。也可以直接将FFMPEG解码输出AVFrame中的data赋值过来
    int32_t linesize[LIBV_NUM_DATA_POINTERS];
}FRAME_PIC_DATA;

//库初始化函数，主要是初始化日志，ffmpeg库等。
void libv_init(LIBV_LOG_FUN log_cbfun, LIBV_LOG_LEVEL_E level, const char* log_file, int printf_flg);
uint8_t * libv_get_version();
FRAME_PIC_DATA *libv_alloc_frame_pic_mem(PICTURE_FORMAT pic_format, int32_t width, int32_t height);
FRAME_PIC_DATA *libv_alloc_frame_pic(PICTURE_FORMAT pic_format, int32_t width, int32_t height);
FRAME_PIC_DATA *libv_copy_new_frame_pic(FRAME_PIC_DATA *psrc_pic);
void libv_free_frame_pic(FRAME_PIC_DATA * frame_pic);

PICTURE_FORMAT libv_get_pic_fmt_by_ffm_fmt(int32_t fmt);
int32_t libv_get_ffm_fmt_by_pic_fmt(PICTURE_FORMAT fmt);

#ifdef __cplusplus
}
#endif

#endif /* LIBVIDEO_H_ */
