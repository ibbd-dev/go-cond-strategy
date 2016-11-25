package servicesDegrade

import (
	"fmt"
	"testing"
	"time"
)

func init() {
	updateMetricFunc = func() {}
	updateMetricDuration = time.Second
	time.Sleep(time.Second * 2)
}

func TestDegrade(t *testing.T) {
	// 初始化事件
	eventName := "access"
	InitEvent(eventName)

	// 配置降级策略
	stragy := &TConf{}
	var hello int
	stragy.Check = func(metric *TEventsMetric) bool {
		if metric.Events[eventName].OneMinute.Count > 100 {
			return true
		}
		return false
	}
	stragy.YesAction = func() {
		hello = 1
	}
	stragy.NoAction = func() {
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
		t.Fatalf("error count")
	}

	// 更新指标
	time.Sleep(time.Second)
	updateMetric()
	println("updateMetric first:")
	fmt.Printf("%+v\n", eventsMetric.Events[eventName].OneMinute)
	fmt.Printf("%+v\n", eventsMetric.Events[eventName].FiveMinute)
	fmt.Printf("%+v\n\n", events[eventName].oneMinute)
	if events[eventName].oneMinute.count != 0 {
		t.Fatalf("error events count")
	}

	if hello != 1 {
		t.Fatalf("error")
	}

	// 配置新的降级策略
	stragy = &TConf{}
	var world int
	hello = 0
	stragy.Check = func(metric *TEventsMetric) bool {
		if metric.Events[eventName].OneMinute.Count > 200 {
			return true
		}
		return false
	}
	stragy.YesAction = func() {
		world = 1
	}
	stragy.NoAction = func() {
		world = 2
	}

	// 增加策略
	AddConf(stragy)

	for i := 0; i < 150; i++ {
		ev := BeginEvent(eventName)
		ev.End()
	}

	// 更新指标
	println("")
	fmt.Printf("%+v\n", events[eventName].oneMinute)
	fmt.Printf("%+v\n", eventsMetric.Events[eventName].OneMinute)
	time.Sleep(time.Second)
	updateMetric()
	println("updateMetric second:")
	fmt.Printf("%+v\n", eventsMetric.Events[eventName].OneMinute)
	fmt.Printf("%+v\n", eventsMetric.Events[eventName].FiveMinute)
	fmt.Printf("%+v\n\n", events[eventName].oneMinute)

	if hello != 1 {
		t.Fatalf("error hello")
	}

	if world != 2 {
		fmt.Printf("%+v\n", eventsMetric.Events[eventName].OneMinute)
		fmt.Printf("%+v\n", eventsMetric.Events[eventName].FiveMinute)
		t.Fatalf("error world")
	}

}
