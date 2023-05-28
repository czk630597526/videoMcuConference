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
    NALU_TYPE_UNDEFINE = 0,                    // 0��δ�涨
    NALU_TYPE_SLICE    = 1 ,                   // 1����IDRͼ���в��������ݻ��ֵ�Ƭ��
    NALU_TYPE_DPA      = 2 ,                   // 2����IDRͼ����A�����ݻ���Ƭ��
    NALU_TYPE_DPB      = 3 ,                   // 3����IDRͼ����B�����ݻ���Ƭ��
    NALU_TYPE_DPC      = 4 ,                   // 4����IDRͼ����C�����ݻ���Ƭ��
    NALU_TYPE_IDR      = 5 ,                   // 5��IDRͼ���Ƭ��
    NALU_TYPE_SEI      = 6 ,                   // 6��������ǿ��Ϣ (SEI)
    NALU_TYPE_SPS      = 7 ,                   // 7�����в�����
    NALU_TYPE_PPS      = 8 ,                   // 8��ͼ�������
    NALU_TYPE_AUD      = 9 ,                   // 9���ָ��
    NALU_TYPE_EOSEQ    = 10,                   // 10�����н�����
    NALU_TYPE_EOSTREAM = 11,                   // 11����������
    NALU_TYPE_FILL     = 12,                   // 12���������
                                               //13 �C 23������
                                               //24 �C 31��Ӧ��������ְ�
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
    uint32_t used_flag;         //�Ƿ�ʹ�ã�0��δʹ�ã�1��ʹ�á��ϲ������ʹ�ô��ֶΡ�
    NALU_TYPE nal_type;          //nal���ͣ�NALU_TYPEö��
    FRAME_TYPE frame_type;       //֡�����ͣ���nal�������ǲ�ͬ�ģ�ָ����I,P,B֡��
    uint32_t  timestamp;         //nal��ʱ��
    uint32_t data_size;         //nal���Ĵ�С
    uint8_t *data_buf;        //nal��ָ�룬ָ��һ����ʱ����������֧�ֶ��̣߳������Ҫ�洢��Ҫ�������ڴ档
                                    //��ʽ��[0,0,0,1]Start sequence + [nal header, data]NAL data
    uint32_t frame_end_flg;      //֡������־λ������һ��nalһ��slice�������ǵ���sliceģʽ�����ڱ�ʶ֡������
                                 //����ʱ������ϲ㲻�����ã�
}H264_NAL_DATA;

typedef struct __H264_RTP_PACKET
{
    uint32_t used_flag;         //�Ƿ�ʹ�ã�0��δʹ�ã�1��ʹ�á��ϲ������ʹ�ô��ֶΡ�
    uint32_t seq;               //rtp�������
    uint32_t data_size;         //rtp���Ĵ�С
    uint8_t *data_buf;          //rtp��ָ�룬ָ��һ����ʱ����������֧�ֶ��̣߳������Ҫ�洢��Ҫ�������ڴ档
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

//����ԭʼͼ�����ݣ�����yuv,���������
//����ʹ�ú���libvideo_alloc_frame_pic���룬������Ҫ����libvideo_free_frame_pic�����ͷš�
//Ҳ�����Լ�ֱ�ӹ�������ṹ�壬data��linesize�ý����������
typedef struct __FRAME_PIC_DATA
{
    PICTURE_FORMAT pic_format;      //��ʾ�洢��ʽ��Ŀǰֻ֧��420p
    PICTURE_TYPE pic_type;          //ͼƬ֡��ʽ���ǣ�i��p��b������������������
    uint32_t  timestamp;         //ͼƬ��ʱ��
    int32_t   key_frame;            //�ͷ�ʱ�ؼ�֡
    int32_t   quality;              //ͼƬ����
    int32_t   width;
    int32_t   height;
    int32_t _alloc_flg;             //data�е��ڴ��Ƿ��ǿ�������ġ��ͷ�ʱ������˱�־Ϊ1�����ͷ�data���ϲ㲻Ҫʹ��
    uint8_t *data[LIBV_NUM_DATA_POINTERS];      //����ʹ��ȥ�����ڴ�,������av_image_alloc���롣Ҳ����ֱ�ӽ�FFMPEG�������AVFrame�е�data��ֵ����
    int32_t linesize[LIBV_NUM_DATA_POINTERS];
}FRAME_PIC_DATA;

//���ʼ����������Ҫ�ǳ�ʼ����־��ffmpeg��ȡ�
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
