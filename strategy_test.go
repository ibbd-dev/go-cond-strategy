package crondStrategy

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
	AddEvent(eventName)

	// 配置策略
	var hello int
	strategy := NewStrategy(func(m *TEventsMetric) TLevel {
		println("第1个策略...")
		if m.Events[eventName].OneMinute.Count > 300 {
			return StatusYes
		}
		return StatusNo

	}, func(status TLevel) {
		if status == StatusYes {
			hello = 1
		} else {
			hello = 2
		}
	})

	// 模拟数据统计数据
	for i := 0; i < 400; i++ {
		ev := BeginEvent(eventName)
		ev.End()
	}

	//fmt.Printf("%+v\n", events[eventName].oneMinute)
	if events[eventName].oneMinute.count != 400 {
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

	// 配置新的策略
	hello = 0
	var world TLevel
	strategy2 := NewStrategy(func(m *TEventsMetric) TLevel {
		println("第2个策略...")
		if m.Events[eventName].OneMinute.Count > 500 {
			return Level5
		} else if m.Events[eventName].OneMinute.Count > 400 {
			return Level4
		} else if m.Events[eventName].OneMinute.Count > 300 {
			return Level3
		} else if m.Events[eventName].OneMinute.Count > 200 {
			return Level2
		}
		return Level1

	}, func(level TLevel) {
		world = level
	})

	for i := 0; i < 250; i++ {
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

	if hello != 2 {
		t.Fatalf("error hello: %d", hello)
	}

	if world != Level2 {
		fmt.Printf("%+v\n", eventsMetric.Events[eventName].OneMinute)
		fmt.Printf("%+v\n", eventsMetric.Events[eventName].FiveMinute)
		t.Fatalf("error world: %d", world)
	}

	// 测试策略的开关
	println("\n测试策略的开关: 停止第1个策略")
	strategy.Stop()
	hello = 0
	world = levelNotInit

	for i := 0; i < 350; i++ {
		ev := BeginEvent(eventName)
		ev.End()
	}
	time.Sleep(time.Second)
	updateMetric()

	if hello != 0 {
		t.Fatalf("error hello: %d", hello)
	}

	if world != Level3 {
		t.Fatalf("error world: %d", world)
	}

	_ = strategy
	_ = strategy2
}

func BenchmarkEvent(b *testing.B) {
	eventName := "access"
	AddEvent(eventName)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ev := BeginEvent(eventName)
			ev.End()
		}
	})
}
