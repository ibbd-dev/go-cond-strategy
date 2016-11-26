# go-cond-strategy

Golang条件策略框架及服务降级框架，根据相关指标，自动对相关的服务进行开启关闭，或者降级，例如原来按100%写日志，降级为50%的概率写日志。

服务降级之后，当条件不再满足的时候，应该能自动恢复。

## Install 

```sh
# 条件策略框架
go get -u github.com/ibbd-dev/go-cond-strategy
```

## 主要概念

- **事件**：例如将从接收到请求到返回数据，定义为一个访问事件，使用时需要先对事件的名字进行初始化。事件分为两类：
  - **计数事件**：只需要统计发生次数的事件，例如错误的发生次数。对应CountEvent函数
  - **耗时事件**：这类时间是有开始和结束之分，可以统计次数和耗时。对应BeginEvent和End函数
- **指标**：指标都含有一个时间段的特征，我们实现的是1分钟和5分钟的指标，具体指标又分为：
  - **次数Count**：事件在统计时间段内发生的次数
  - **平均耗时AvgTime**：每次事件的平均消耗事件，耗时事件才有该指标
- **策略**：当指标满足某些条件时，可以执行相应的动作，否则可以执行另外一组动作，就是策略。例如当某事件的平均耗时很高时，系统能自动进行降级，或者当某事件的发生次数达到某个阀值的时候，自动发送邮件等。

策略Check函数可以返回一个级别（定义了5个可用等级, 从Level1到Level5），当级别发生改变的时候，就会执行Action动作。例如，我们可以将系统的压力分成5个级别，外部可以定义在什么级别开启什么服务或者关闭什么服务等。

如果我们只使用其中的两个级别（也定义两个常量：StatusYes, StatusNo），就可以变成true or false的策略。

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

    "github.com/ibbd-dev/go-cond-strategy"
)

func main() {
	// 初始化事件
	eventName := "access"
	condStrategy.InitEvent(eventName)


	// 配置策略
	var hello int
	strategy := condStrategy.NewStrategy(func(m *condStrategy.TEventsMetric) condStrategy.TLevel {
		println("第1个策略...")
		if m.Events[eventName].OneMinute.Count > 300 {
			return StatusYes
		}
		return StatusNo

	}, func(status condStrategy.TLevel) {
		if status == condStrategy.StatusYes {
			hello = 1  // Do somethings...
		} else {
			hello = 2  // Do somethings else...
		}
	})

	// 模拟数据统计数据
	for i := 0; i < 400; i++ {
		ev := BeginEvent(eventName)
		ev.End()
	}

	// 等待指标的更新
	time.Sleep(time.Minute)
	time.Sleep(time.Second)

	// 更新指标
	if hello != 1 {
		println("error")
	}
}
```

## 性能数据

BeginEvent和End的性能数据

```
// 同时记录1分钟数据和5分钟数据
BenchmarkEvent-4    20000000            73.6 ns/op

// 只记录1分钟数据，5分钟数据在后续统计时补上
BenchmarkEvent-4    20000000            61.7 ns/op

```



