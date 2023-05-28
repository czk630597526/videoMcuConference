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
    H264_DECODER_ID _id;     //�������ı�ʶ���ϲ㲻�ô���
    void *_private_data;     //   ÿ����������˽�����ݽṹ�壬�ϲ㲻���޸ġ�

    //���봦������
    /*
    pdecth��     �������ľ����
    pout_pic��������ݣ����е�data�ϲ㲻�����룬Ҳ�����ͷţ�decoder�Ḻ��ά���� �����data����һ�ε���decode_video����
                                     ǰ���ǿ��õģ������Ҫ��������������Ҫ����libvideo_copy_frame_pic���п�������������Ŀ����Ϊ�˼����ڴ濽����
                                     �����ʱ��ȡ���������ʱ������Ϊ�����֡�������������֡��
    got_picture_ptr:�ͷ���֡�����Ϊ0��ʾ�˴�û��֡�����1��ʾ��֡�����
    pnal    ���������ݣ����Я��ʱ�����������֡���ʱ��ʱ����ֵ��
    return  : ʧ��Ϊ-1������0��ʾ�����byte����0��ʾû�н��롣
     */
    int32_t (*decode_video)(H264_DECODER_TH *pdecth, FRAME_PIC_DATA *pout_pic, int32_t *got_picture_ptr, const H264_NAL_DATA *pnal);

};

//���������ṹ�壬�ο����ģʽ����������.?��Щ��ȷ��ģʽ�����Ƿ�й���
typedef struct __H264_DECODER_FACTORY
{
    H264_DECODER_ID id;     //�������ı�ʶ
    int32_t (*init)(H264_DECODER_TH *pdecth);      //��ʼ��������pdecth�������ϲ����롣
    void (*release)(H264_DECODER_TH *pdecth);   //�ͷź���������������������pdecth�������ϲ��ͷš�

}H264_DECODER_FACTORY;

void h264_decoder_init();   //�ϲ㲻�õ��ã�����libvideo_init�е��á�
int32_t h264_decoder_register(H264_DECODER_FACTORY *pdec_fact);//�ڲ�ʹ�ã��ϲ㲻�õ���

//���ݽ�����id����һ������������Ҫ����free_decoder�Ƿ�
H264_DECODER_TH *alloc_decoder_by_id(H264_DECODER_ID id);
void free_decoder(H264_DECODER_TH *pdecth);


#ifdef __cplusplus
}
#endif

#endif /* H264_DECODER_C_ */
