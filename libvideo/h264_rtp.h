/*
 * h264_rtp.h
 *
 */

#ifndef H264_RTP_H
#define H264_RTP_H

#include <stdint.h>
#include "libv_log.h"
#include "libvideo.h"
/*
 * �˿��ݲ�֧�ֶ��߳�
 * */

#ifdef __cplusplus
extern "C" {
#endif

typedef enum
{
    PACKET_MODE_SINGLE = 0, //������ģʽ
    PACKET_MODE_NON_INTERLEAVED = 1,    //�ǽ���ʽ

}PACKET_MODE_E;


typedef void * H264_RTP_DEC_TH;

H264_RTP_DEC_TH alloc_h264_rtp_dec_th();
void free_h264_rtp_dec_th(H264_RTP_DEC_TH  pdec_th);
//rtp_buf:����rtp��ͷ.����ֵ����ð��ĸ���.-1��ʾ����ʧ��
int h264_dec_rtp(H264_RTP_DEC_TH  pdec_th, uint8_t *rtp_buf, int buf_len);

//�����������ݵ�ָ�룬û�������򷵻ؿա��������Ҫһֱ���ã�ֱ������nullΪֹ���������ݻᱻ����
//���ص�ָ���ָ���еĻ�������������Ϊ�ϲ�Ĵ洢�������������Ҫ����һ��ʱ�䣬����Ҫ���������µĻ�������������
//used_flag�ֶη���ǰ�ᱻ��Ϊ0.
H264_NAL_DATA * h264_dec_rtp_get_nal_data(H264_RTP_DEC_TH  pdec_th);

typedef void * H264_RTP_ENC_TH;
/*
 *pload_type:�������ͣ�s_seq����ʼ���к�,ʹ���������Ϊ��ʼ��
 *ssrc��ͬ��Դid;frame_rate��֡�ʣ�Ŀǰ��֧��������������ܻ�֧�ָ�����
 *mtu:����ip����������ƣ�����1500����ֵ����ip��ͷ��udp��ͷ��rtpͷ�ĳ��ȡ�
 * */
H264_RTP_ENC_TH * alloc_h264_rtp_enc_th(uint8_t pload_type, uint16_t s_seq, uint32_t ssrc, uint32_t mtu, PACKET_MODE_E pm);
void free_h264_rtp_enc_th(H264_RTP_ENC_TH  ph264dec_th);
//nal_buf��һ��NAL��Ԫ���ݣ�����������ֽ�[0,0,0,1]�����ֽ�[0,0,1]����ʼ��
//timestamp:��NAL��ʱ������Ҫ�ϲ㴫�ݣ���Ϊδ���������ġ�
//�ɹ����أ������rtp��������ʧ�ܷ��أ�-1��
int32_t h264_enc_rtp_one_nalu(H264_RTP_ENC_TH ph264dec_th, uint8_t *nal_buf, int32_t buf_len, uint32_t timestamp, uint32_t mark_flg);

//an_buf��һ��access unit��Ԫ���ݣ�����������ֽ�[0,0,0,1]�����ֽ�[0,0,1]����ʼ��.
//timestamp:��NAL��ʱ������Ҫ�ϲ㴫�ݣ���Ϊδ���������ġ�
//�ɹ����أ������rtp��������ʧ�ܷ��أ�-1��
int32_t h264_enc_rtp_access_unit(H264_RTP_ENC_TH ph264enc_th, uint8_t *an_buf, int32_t buf_len, uint32_t timestamp);

//�����������ݵ�ָ�룬û�������򷵻ؿա��������Ҫһֱ���ã�ֱ������nullΪֹ���������ݻᱻ����
//���ص�ָ���ָ���еĻ�������������Ϊ�ϲ�Ĵ洢�������������Ҫ����һ��ʱ�䣬����Ҫ���������µĻ�������������
//used_flag�ֶη���ǰ�ᱻ��Ϊ0.
H264_RTP_PACKET * h264_enc_rtp_get_rtp_data(H264_RTP_ENC_TH  penc_th);

#ifdef __cplusplus
}
#endif



#endif

