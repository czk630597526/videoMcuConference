/*
 * h264_encoder.h
 *
 */

#ifndef H264_ENCODER_H_
#define H264_ENCODER_H_

#include <stdint.h>
#include "libv_log.h"
#include "libvideo.h"

#ifdef __cplusplus
extern "C" {
#endif

#define      MAX_NAL_UNITS_COUNT_IN_LAYER      512

typedef enum __H264_ENCODER_ID
{
    H264_ENCODER_ID_OPENH264 = 0,//目前解码只需要OPENH264
    H264_ENCODER_ID_MAX = 1,
}H264_ENCODER_ID;

//输出slice模式，目前支持单slice和动态slice（根据slice大小调整个数）
typedef enum {
    H264_SM_SINGLE_SLICE         = 0, //   | SliceNum==1
    H264_SM_FIXEDSLCNUM_SLICE    = 1, //   | according to SliceNum     | Enabled dynamic slicing for multi-thread
    H264_SM_RASTER_SLICE         = 2, //   | according to SlicesAssign | Need input of MB numbers each slice. In addition, if other constraint in SSliceArgument is presented, need to follow the constraints. Typically if MB num and slice size are both constrained, re-encoding may be involved.
//  H264_SM_ROWMB_SLICE          = 3, //   | according to PictureMBHeight  | Typical of single row of mbs each slice?+ slice size constraint which including re-encoding
    H264_SM_DYN_SLICE            = 4, //此模式目前bug较多，在1080p是，必须是 max_slice_size对于1800，  | according to SliceSize    | Dynamic slicing (have no idea about slice_nums until encoding current frame)
    H264_SM_AUTO_SLICE           = 5, //   | according to thread number，跟随线程个数
//  H264_SM_RESERVED             = 6
}H264_SLICE_MODE_E;

//SLICE配置信息
typedef struct __SLICE_CONFIG
{
    uint32_t  max_slice_size;    //slice大小最大值，H264_SM_DYN_SLICE是有效
    uint32_t  slice_num;        //slice个数，H264_SM_FIXEDSLCNUM_SLICE时有效
    uint32_t  slice_mb_num;      //每个slice的宏块（16*16的视频单元）的个数，H264_SM_RASTER_SLICE是有效
}SLICE_CONFIG;

typedef enum {
    H264_LOW_COMPLEXITY,             ///< the lowest compleixty,the fastest speed,
    H264_MEDIUM_COMPLEXITY,          ///< medium complexity, medium speed,medium quality
    H264_HIGH_COMPLEXITY             ///< high complexity, lowest speed, high quality
} H264_ECOMPLEXITY_MODE;


typedef struct __ENCODE_PARAMETER
{
    PICTURE_FORMAT pic_format;      //表示存储格式，目前只支持420p
    int32_t  profile_id;    //  这个字段暂时不用。因为目前只支持base。base延迟地，适合我们的应用。
    int32_t  width;
    int32_t  height;
    int32_t  frame_rate;    //指定帧率,不知道填写0；
    int32_t  multiple_thread_id;  //目前填写1，openh264不支持多个 1 # 0: auto(dynamic imp. internal encoder); 1: multiple threads imp. disabled; > 1: count number of threads;
//    int32_t  max_qp;    //质量最大值，越高质量越好。建议：51:目前这三个值没有效果，所以暂时不开放
//    int32_t  min_qp;    //质量最小值     建议：0
//    int32_t  qp;        //当前质量      建议：26
    int32_t  bitrate;        //目标比特率。单位为bps
    int32_t  max_bitrate;        //最大比特率。单位为bps
    H264_SLICE_MODE_E slice_mode;  //slice模式
    SLICE_CONFIG  slice_cfg;       //当slice_mode为H264_SM_DYN_SLICE时有效，指定slice的最大值
    H264_ECOMPLEXITY_MODE complexity_mode;  //编码复杂度。
}ENCODE_PARAMETER;

/* Bitstream inforamtion of a layer being encoded */
typedef struct __LAYER_BS_INFO
{
    uint8_t tmp_id;
    FRAME_TYPE frame_type;       //帧的类型，和nal的类型是不同的，指明是I,P,B帧等

    int32_t  nal_count;                  // Count number of nal_data already
    H264_NAL_DATA  nal_data[MAX_NAL_UNITS_COUNT_IN_LAYER];
} LAYER_BS_INFO;


typedef struct __H264_ENCODER_TH    H264_ENCODER_TH;
struct __H264_ENCODER_TH
{
    H264_ENCODER_ID _id;     //编码器的标识，上层不用处理。
    ENCODE_PARAMETER param;
    void *_private_data;     //   每个编码器的私有数据结构体，上层不能修改。

    //编码处理函数：
    /*
     * pencth:解码器句柄
     * pin_pic：输入图片
     * player_bsi：输出nal数据
     */
    int32_t (*encode_frame)(H264_ENCODER_TH *pencth, FRAME_PIC_DATA *pin_pic, LAYER_BS_INFO *player_bsi);

    //获取sps，pps数据。
    int32_t (*encode_parameter_sets)(H264_ENCODER_TH *pencth, LAYER_BS_INFO *player_bsi);
    //强制输出I帧，is_idr表示是否是IDR.设置IDR标志后，会先输出sps和pps。
    int32_t (*force_intra_frame) (H264_ENCODER_TH *pencth, int32_t  is_idr);
    //重置帧率，一般用在开始无法知道实际帧率的情况下。
    void (*reset_frame_rate)(H264_ENCODER_TH *pencth, int32_t  frame_rate);
};

//工厂方法结构体，参考设计模式：工厂方法.?有些不确定模式名称是否叫工厂
typedef struct __H264_ENCODER_FACTORY
{
    H264_ENCODER_ID id;     //解码器的标识
    int32_t (*init)(H264_ENCODER_TH *pencth, ENCODE_PARAMETER *pparam);      //初始化函数，pdecth对象由上层申请。
    void (*release)(H264_ENCODER_TH *pencth);   //释放函数，类似于析构函数，pdecth对象由上层释放。

}H264_ENCODER_FACTORY;

void h264_encoder_init();   //上层不用调用，会在libvideo_init中调用。
int32_t h264_encoder_register(H264_ENCODER_FACTORY *penc_fact);//内部使用函数。

//根据解码器id申请一个编码器，需要调用free_encoder是否。
H264_ENCODER_TH *alloc_encoder_by_id(H264_ENCODER_ID id, ENCODE_PARAMETER *pparam);
void free_encoder(H264_ENCODER_TH *pencth);



#ifdef __cplusplus
}
#endif


#endif /* H264_ENCODER_H_ */
