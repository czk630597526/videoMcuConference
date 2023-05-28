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
    H264_ENCODER_ID_OPENH264 = 0,//Ŀǰ����ֻ��ҪOPENH264
    H264_ENCODER_ID_MAX = 1,
}H264_ENCODER_ID;

//���sliceģʽ��Ŀǰ֧�ֵ�slice�Ͷ�̬slice������slice��С����������
typedef enum {
    H264_SM_SINGLE_SLICE         = 0, //   | SliceNum==1
    H264_SM_FIXEDSLCNUM_SLICE    = 1, //   | according to SliceNum     | Enabled dynamic slicing for multi-thread
    H264_SM_RASTER_SLICE         = 2, //   | according to SlicesAssign | Need input of MB numbers each slice. In addition, if other constraint in SSliceArgument is presented, need to follow the constraints. Typically if MB num and slice size are both constrained, re-encoding may be involved.
//  H264_SM_ROWMB_SLICE          = 3, //   | according to PictureMBHeight  | Typical of single row of mbs each slice?+ slice size constraint which including re-encoding
    H264_SM_DYN_SLICE            = 4, //��ģʽĿǰbug�϶࣬��1080p�ǣ������� max_slice_size����1800��  | according to SliceSize    | Dynamic slicing (have no idea about slice_nums until encoding current frame)
    H264_SM_AUTO_SLICE           = 5, //   | according to thread number�������̸߳���
//  H264_SM_RESERVED             = 6
}H264_SLICE_MODE_E;

//SLICE������Ϣ
typedef struct __SLICE_CONFIG
{
    uint32_t  max_slice_size;    //slice��С���ֵ��H264_SM_DYN_SLICE����Ч
    uint32_t  slice_num;        //slice������H264_SM_FIXEDSLCNUM_SLICEʱ��Ч
    uint32_t  slice_mb_num;      //ÿ��slice�ĺ�飨16*16����Ƶ��Ԫ���ĸ�����H264_SM_RASTER_SLICE����Ч
}SLICE_CONFIG;

typedef enum {
    H264_LOW_COMPLEXITY,             ///< the lowest compleixty,the fastest speed,
    H264_MEDIUM_COMPLEXITY,          ///< medium complexity, medium speed,medium quality
    H264_HIGH_COMPLEXITY             ///< high complexity, lowest speed, high quality
} H264_ECOMPLEXITY_MODE;


typedef struct __ENCODE_PARAMETER
{
    PICTURE_FORMAT pic_format;      //��ʾ�洢��ʽ��Ŀǰֻ֧��420p
    int32_t  profile_id;    //  ����ֶ���ʱ���á���ΪĿǰֻ֧��base��base�ӳٵأ��ʺ����ǵ�Ӧ�á�
    int32_t  width;
    int32_t  height;
    int32_t  frame_rate;    //ָ��֡��,��֪����д0��
    int32_t  multiple_thread_id;  //Ŀǰ��д1��openh264��֧�ֶ�� 1 # 0: auto(dynamic imp. internal encoder); 1: multiple threads imp. disabled; > 1: count number of threads;
//    int32_t  max_qp;    //�������ֵ��Խ������Խ�á����飺51:Ŀǰ������ֵû��Ч����������ʱ������
//    int32_t  min_qp;    //������Сֵ     ���飺0
//    int32_t  qp;        //��ǰ����      ���飺26
    int32_t  bitrate;        //Ŀ������ʡ���λΪbps
    int32_t  max_bitrate;        //�������ʡ���λΪbps
    H264_SLICE_MODE_E slice_mode;  //sliceģʽ
    SLICE_CONFIG  slice_cfg;       //��slice_modeΪH264_SM_DYN_SLICEʱ��Ч��ָ��slice�����ֵ
    H264_ECOMPLEXITY_MODE complexity_mode;  //���븴�Ӷȡ�
}ENCODE_PARAMETER;

/* Bitstream inforamtion of a layer being encoded */
typedef struct __LAYER_BS_INFO
{
    uint8_t tmp_id;
    FRAME_TYPE frame_type;       //֡�����ͣ���nal�������ǲ�ͬ�ģ�ָ����I,P,B֡��

    int32_t  nal_count;                  // Count number of nal_data already
    H264_NAL_DATA  nal_data[MAX_NAL_UNITS_COUNT_IN_LAYER];
} LAYER_BS_INFO;


typedef struct __H264_ENCODER_TH    H264_ENCODER_TH;
struct __H264_ENCODER_TH
{
    H264_ENCODER_ID _id;     //�������ı�ʶ���ϲ㲻�ô���
    ENCODE_PARAMETER param;
    void *_private_data;     //   ÿ����������˽�����ݽṹ�壬�ϲ㲻���޸ġ�

    //���봦������
    /*
     * pencth:���������
     * pin_pic������ͼƬ
     * player_bsi�����nal����
     */
    int32_t (*encode_frame)(H264_ENCODER_TH *pencth, FRAME_PIC_DATA *pin_pic, LAYER_BS_INFO *player_bsi);

    //��ȡsps��pps���ݡ�
    int32_t (*encode_parameter_sets)(H264_ENCODER_TH *pencth, LAYER_BS_INFO *player_bsi);
    //ǿ�����I֡��is_idr��ʾ�Ƿ���IDR.����IDR��־�󣬻������sps��pps��
    int32_t (*force_intra_frame) (H264_ENCODER_TH *pencth, int32_t  is_idr);
    //����֡�ʣ�һ�����ڿ�ʼ�޷�֪��ʵ��֡�ʵ�����¡�
    void (*reset_frame_rate)(H264_ENCODER_TH *pencth, int32_t  frame_rate);
};

//���������ṹ�壬�ο����ģʽ����������.?��Щ��ȷ��ģʽ�����Ƿ�й���
typedef struct __H264_ENCODER_FACTORY
{
    H264_ENCODER_ID id;     //�������ı�ʶ
    int32_t (*init)(H264_ENCODER_TH *pencth, ENCODE_PARAMETER *pparam);      //��ʼ��������pdecth�������ϲ����롣
    void (*release)(H264_ENCODER_TH *pencth);   //�ͷź���������������������pdecth�������ϲ��ͷš�

}H264_ENCODER_FACTORY;

void h264_encoder_init();   //�ϲ㲻�õ��ã�����libvideo_init�е��á�
int32_t h264_encoder_register(H264_ENCODER_FACTORY *penc_fact);//�ڲ�ʹ�ú�����

//���ݽ�����id����һ������������Ҫ����free_encoder�Ƿ�
H264_ENCODER_TH *alloc_encoder_by_id(H264_ENCODER_ID id, ENCODE_PARAMETER *pparam);
void free_encoder(H264_ENCODER_TH *pencth);



#ifdef __cplusplus
}
#endif


#endif /* H264_ENCODER_H_ */
