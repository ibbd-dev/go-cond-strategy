# go-services-degrade

Golang服务降级框架，根据相关指标，自动对相关的服务进行开启关闭，或者降级，例如原来按100%写日志，降级为50%的概率写日志。

服务降级之后，当条件不再满足的时候，应该能自动恢复。

## Install 

```sh
go get -u github.com/ibbd-dev/go-services-degrade
```


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



