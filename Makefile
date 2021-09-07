clean:
	rm -f *.out
	rm -f *.test
	rm -f *.svg

bench:
	go test -bench=. -benchmem
cover:
	go test -cover

profile:
	go test -run=NONE -bench=BenchmarkRender -cpuprofile=cpu.out
	go test -run=NONE -bench=BenchmarkRender -memprofile=mem.out
	go tool pprof -svg -output cpu.svg ./goldmark-gemtext.test cpu.out
	go tool pprof -svg -output mem.svg ./goldmark-gemtext.test mem.out
