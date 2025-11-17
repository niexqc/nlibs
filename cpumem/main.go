package main

import (
	"flag"
	"fmt"
	"log/slog"

	cpulaoder "github.com/niexqc/nlibs/cpumem/cpuloader"
	"github.com/niexqc/nlibs/cpumem/memloader"

	"github.com/niexqc/nlibs/ntools"
)

func main() {
	ntools.SlogConf("d", "debug", 1, 2)
	cpu := 40.0
	mem := 60.0
	chktm := int64(3)

	cpuavgtm := int64(1)

	flag.Float64Var(&cpu, "cpu", cpu, fmt.Sprintf("Cpu目标负载:%f", cpu))
	flag.Float64Var(&mem, "mem", mem, fmt.Sprintf("Mem目标负载:默认值%f", mem))
	flag.Int64Var(&chktm, "chktm", chktm, fmt.Sprintf("检查间隔时间:默认%d秒", chktm))
	flag.Int64Var(&cpuavgtm, "cpuavgtm", cpuavgtm, fmt.Sprintf("Cpu平均值计算时间:默认%d秒", cpuavgtm))
	flag.Parse()

	slog.Info(fmt.Sprintf("Cpu目标负载:%.2f,Mem目标负载:%.2f,检查间隔时间:%d秒,Cpu平均值计算时间:%d秒", cpu, mem, chktm, cpuavgtm))

	go cpulaoder.NewCpuLoader(cpu, chktm, cpuavgtm).Start()
	memloader.NewMemLoader(mem, chktm).Start()
}
