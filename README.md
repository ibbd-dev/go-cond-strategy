# go-services-degrade

Golang服务降级框架，根据相关指标，自动对相关的服务进行开启关闭，或者降级，例如原来按100%写日志，降级为50%的概率写日志。

服务降级之后，当条件不再满足的时候，应该能自动恢复。

## Install 

```sh
go get -u github.com/ibbd-dev/go-services-degrade
```

## 主要概念

- 事件：例如将从接收到请求到返回数据，定义为一个访问事件，使用时需要先对事件的名字进行初始化。事件分为两类：
  - 计数事件：只需要统计发生次数的事件，例如错误的发生次数。对应CountEvent函数
  - 耗时事件：这类时间是有开始和结束之分，可以统计次数和耗时。对应BeginEvent和End函数
- 指标：指标都含有一个时间段的特征，我们实现的是1分钟和5分钟的指标，具体指标又分为：
  - 次数Count：事件在统计时间段内发生的次数
  - 平均耗时AvgTime：每次事件的平均消耗事件，耗时事件才有该指标
- 策略：当指标满足某些条件时，可以执行相应的动作，否则可以执行另外一组动作，就是策略。例如当某事件的平均耗时很高时，系统能自动进行降级，或者当某事件的发生次数达到某个阀值的时候，自动发送邮件等。

## 实现思路

指标：

- 事件总次数：一段时间内的事件发生的总次数
- 事件平均耗时指标：一段时间内完成事件耗时的均值

时间段：

- 1分钟
- 5分钟

前端可以设置相关指标满足什么条件时，触发什么操作。后台就会周期性计算是否满足触发条件，减少判断的耗时。

## Example

```go
package main

import (
	"fmt"
	"time"
)

func init() {
	// 模拟测试
	updateMetricFunc = func() {}
	updateMetricDuration = time.Second
	time.Sleep(time.Second * 2)
}

func main() {
	// 初始化事件
	eventName := "access"
	InitEvent(eventName)

	// 配置降级策略
	stragy := &TConf{}
	var hello int
	stragy.Check = func(metric *TEventsMetric) bool {
		if metric.Events[eventName].OneMinute.Count > 100 {
			// 一分钟内次数超过100，则触发该条件 
			return true
		}
		return false
	}
	stragy.YesAction = func() {
		// 被触发之后需要执行的代码
		hello = 1
	}
	stragy.NoAction = func() {
		// 不被触发时需要执行的代码
		hello = 2
	}

	// 增加策略
	AddConf(stragy)

	// 模拟数据统计数据
	for i := 0; i < 200; i++ {
		ev := BeginEvent(eventName)
		ev.End()
	}

	fmt.Printf("%+v\n", events[eventName].oneMinute)
	if events[eventName].oneMinute.count != 200 {
		fmt.Printf("%+v\n", events[eventName].oneMinute)
	}

	// 模拟更新指标
	time.Sleep(time.Second)
	updateMetric()
	println("updateMetric first:")
	fmt.Printf("%+v\n", eventsMetric.Events[eventName].OneMinute)
	fmt.Printf("%+v\n", eventsMetric.Events[eventName].FiveMinute)
	fmt.Printf("%+v\n\n", events[eventName].oneMinute)
	if events[eventName].oneMinute.count != 0 {
		fmt.Println("error events count")
	}

	if hello != 1 {
		fmt.Println("error value of hello")
	}
}
```



