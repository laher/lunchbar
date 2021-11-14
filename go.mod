module github.com/laher/lunchbar

go 1.17

require (
	github.com/Kodeworks/golang-image-ico v0.0.0-20141118225523-73f0f4cfade9
	github.com/apex/log v1.9.0
	github.com/getlantern/systray v1.1.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/joho/godotenv v1.4.0
	github.com/laher/lunchbox v0.0.0-20211112085600-b8d31d5220a8
	github.com/matryer/xbar/pkg/plugins v0.0.0-20210918110050-1410be750e94
	golang.org/x/exp v0.0.0-20211105205138-14c72366447f
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d
)

require (
	code.cloudfoundry.org/bytefmt v0.0.0-20211005130812-5bb3c17173e5 // indirect
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d // indirect
	github.com/getlantern/context v0.0.0-20190109183933-c447772a6520 // indirect
	github.com/getlantern/errors v1.0.1 // indirect
	github.com/getlantern/golog v0.0.0-20210606115803-bce9f9fe5a5f // indirect
	github.com/getlantern/hex v0.0.0-20190417191902-c6586a6fe0b7 // indirect
	github.com/getlantern/hidden v0.0.0-20201229170000-e66e7f878730 // indirect
	github.com/getlantern/ops v0.0.0-20200403153110-8476b16edcd6 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/itchyny/gojq v0.12.5 // indirect
	github.com/itchyny/timefmt-go v0.1.3 // indirect
	github.com/leaanthony/go-ansi-parser v1.2.0 // indirect
	github.com/lestrrat-go/strftime v1.0.5 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible // indirect
	github.com/matryer/xbar/pkg/metadata v0.0.0-20210918110050-1410be750e94 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mtibben/androiddnsfix v0.0.0-20200907095054-ff0280446354 // indirect
	github.com/nojima/httpie-go v0.7.0 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/pborman/getopt v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/vbauerster/mpb/v5 v5.4.0 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/sys v0.0.0-20211019181941-9d821ace8654 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	src.elv.sh v0.16.3 // indirect
)

exclude github.com/matryer/xbar/pkg/metadata v0.0.0-00010101000000-000000000000

//replace github.com/matryer/xbar/pkg/plugins => ../xbar/pkg/plugins

replace github.com/matryer/xbar/pkg/plugins => github.com/laher/xbar/pkg/plugins v0.0.0-20210927175547-f69614fb2e13

replace github.com/matryer/xbar/pkg/metadata => github.com/laher/xbar/pkg/metadata v0.0.0-20210927175547-f69614fb2e13

//replace github.com/laher/lunchbox => ../lunchbox
