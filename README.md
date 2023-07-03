# MQTT5-Serial-433Mhz-Remote

串口433Mhz学习型收发器，转MQTT5协议，以接入Home Assistant，实现低成本的车库门控制方案。

## 硬件要求

* 确定要控制器为433Mhz学习码
* 已拥有[该或类似](https://item.taobao.com/item.htm?spm=a1z09.2.0.0.24982e8d872esA&id=530515308004&_u=31juajg00185)串口或USB转串口的信号首发模块
* 已获得自己车库钥匙按钮对应的控制码
* 已部署一个Linux设备，可以为OpenWrt

## 功能

* Home Assistant侧可以观测到传感器当前是否`可用`/`不可用`状态
* `开启`/`关闭`车门后，Home Assistant侧可用观测到`开启中`/`关闭中`等状态
* 支持掉线自动重连
* 仅支持MQTT5协议
* 已对MT7621平台的二进制文件进行压缩，大小仅占1.85MB
* 提供Web API备用通道，防止连接不上MQTT Broker时，被锁在车库外

## 最佳实践

1. 车库安装一个搭载OpenWrt的路由器
2. 接入4G上网卡，或拉一个网线连上网
3. 路由器USB接口插上上述模块
4. 使用本项目

## Home Assistant配置参考

```yaml
mqtt:
  cover:
    - name: "Garage Door"
      unique_id: cover.garage_door
      qos: 0
      retain: false
      command_topic: garage/door/cover/set
      payload_open: OPEN
      payload_close: CLOSE
      payload_stop: STOP
      state_topic: garage/door/cover/state
      state_open: open
      state_opening: opening
      state_closed: closed
      state_closing: closing
      state_stopped: stopped
      availability_topic: garage/door/cover/availability
      payload_available: online
      payload_not_available: offline
```