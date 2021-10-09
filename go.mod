module github.com/laher/lunchbar

go 1.17

require (
	github.com/Kodeworks/golang-image-ico v0.0.0-20141118225523-73f0f4cfade9
	github.com/apex/log v1.9.0
	github.com/getlantern/systray v1.1.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/james-barrow/golang-ipc v0.0.0-20210227130457-95e7cc81f5e2
	github.com/joho/godotenv v1.4.0
	github.com/matryer/xbar/pkg/plugins v0.0.0-20210918110050-1410be750e94
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d
	src.elv.sh v0.16.3
)

require (
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/getlantern/context v0.0.0-20190109183933-c447772a6520 // indirect
	github.com/getlantern/errors v0.0.0-20190325191628-abdb3e3e36f7 // indirect
	github.com/getlantern/golog v0.0.0-20190830074920-4ef2e798c2d7 // indirect
	github.com/getlantern/hex v0.0.0-20190417191902-c6586a6fe0b7 // indirect
	github.com/getlantern/hidden v0.0.0-20190325191715-f02dbb02be55 // indirect
	github.com/getlantern/ops v0.0.0-20190325191751-d70cb0d6f85f // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/leaanthony/go-ansi-parser v1.2.0 // indirect
	github.com/matryer/xbar/pkg/metadata v0.0.0-20210918110050-1410be750e94 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	go.etcd.io/bbolt v1.3.5 // indirect
	golang.org/x/sys v0.0.0-20210820121016-41cdb8703e55 // indirect
)

exclude github.com/matryer/xbar/pkg/metadata v0.0.0-00010101000000-000000000000

//replace github.com/matryer/xbar/pkg/plugins => ../xbar/pkg/plugins

replace github.com/matryer/xbar/pkg/plugins => github.com/laher/xbar/pkg/plugins v0.0.0-20210927175547-f69614fb2e13

replace github.com/matryer/xbar/pkg/metadata => github.com/laher/xbar/pkg/metadata v0.0.0-20210927175547-f69614fb2e13
