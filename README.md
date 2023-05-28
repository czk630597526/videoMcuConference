video mcu conference

本库是一个golang使用ffmpeg库进行混屏的实例，仅仅在linux的环境下进行操作。
cgo是golang中一个复杂但是有趣，并且功能非常强大的特性，可以在某些底层场景使用千锤百炼的c库，在上层使用golang进行业务封装，提高了开发效率，但苦于没有一些简单的教程，因此本人开发了这么一个小demo，用于大家学习与交流

libvideo：
  此包调用了c库，libmedia.so，并且将c函数封装成go可以调用的版本，需要用libmedia的c代码编译出来，代码后面提交，现在可以直接使用lib包里so调用

libvideo_codec:
  一些rtp转nal的基础操作
  
conference：
  做了一些比较简单的业务封装，并且提供了比较高层次的api调用，仅仅操作cfgmgr就能做到一些混屏的功能
  
translayer：
  是将cfgMgr封装成了一个网络调用，根据相应的网络协议就能进行混屏（未经测试，如果使用需要自己写下单元测试）
  
  
  
  在main函数里提供了，提供了一些测试的参考函数，主要是MainConferenceMgr()和MainConference()，提供了cfgMgr的接口测试

本库目前是一个定死为使用h264编解码，固定帧率的demo版本，在各个参会者中收包进行解码，推到会议室里，在会议室里进行混屏和编码，将编完码的数据推往各个参会者，后续继续开发进行协商，协商为参会者需要的视频编解码格式。


run：
  将lib中的库放到你想放的目录下，如/usr/lib/testVideo，将libvideo/libmedia_api中的#cgo LDFLAGS: -L/usr/lib/testVideo ，链接上你存放的目录，然后export LD_LIBRARY_PATH=/usr/lib/testVideo，后面进行go build
