package command

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/hinshun/zap/zapper"
	"github.com/hinshun/zap/zapper/basic"
	"github.com/hinshun/zap/zapper/p2p"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/urfave/cli"
	mpb "github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

func App(ctx context.Context) *cli.App {
	app := cli.NewApp()
	app.Name = "zap"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "log-level,l",
			Usage:  "set the logging level [debug, info, warn, error, fatal, panic]",
			Value:  "debug",
			EnvVar: "LABD_LOG_LEVEL",
		},
		cli.StringFlag{
			Name:  "zapper",
			Usage: "set zapper to transport the file [basic]",
			Value: "basic",
		},
		cli.BoolFlag{
			Name:  "listen",
			Usage: "to zap or to be zapped",
		},
		cli.IntFlag{
			Name:  "port",
			Usage: "specify a port to be zappable at",
			Value: 0,
		},
	}
	app.Action = rootAction

	// Setup tracers and context.
	AttachAppContext(ctx, app)

	return app
}

func AttachAppContext(ctx context.Context, app *cli.App) {
	before := app.Before
	app.Before = func(c *cli.Context) error {
		if before != nil {
			if err := before(c); err != nil {
				return err
			}
		}

		level, err := zerolog.ParseLevel(c.GlobalString("log-level"))
		if err != nil {
			return err
		}


		out := zerolog.ConsoleWriter{Out: os.Stderr}
		rootLogger := zerolog.New(out).Level(level).With().Timestamp().Logger()
		logger := &rootLogger
		ctx = logger.WithContext(ctx)
		c.App.Metadata["context"] = ctx
		return nil
	}
}

func CommandContext(c *cli.Context) context.Context {
	return c.App.Metadata["context"].(context.Context)
}

func rootAction(c *cli.Context) error {
	if c.GlobalBool("listen") {
		return zappedAction(c)
	}
	return zapAction(c)
}

func zapAction(c *cli.Context) error {
	if c.NArg() != 2 {
		return errors.New("must specify src and dst")
	}

	var (
		zpr zapper.Zapper
		err error
	)
	switch c.GlobalString("zapper") {
	case "basic":
		zpr, err = basic.NewZapper()
	case "p2p":
		zpr, err = p2p.NewZapper()
	default:
		return errors.Errorf("unsupported zapper %q", c.GlobalString("zapper"))
	}
	if err != nil {
		return err
	}

	ctx := CommandContext(c)
	src, dst := c.Args().Get(0), c.Args().Get(1)
	return zpr.Zap(ctx, src, dst)
}

func zappedAction(c *cli.Context) error {
	var (
		zpd zapper.Zapped
		err error
	)
	switch c.GlobalString("zapper") {
	case "basic":
		zpd, err = basic.NewZapped()
	case "p2p":
		zpd, err = p2p.NewZapped()
	default:
		return errors.Errorf("unsupported zapper %q", c.GlobalString("zapper"))
	}
	if err != nil {
		return err
	}
	defer zpd.Close()

	ctx := CommandContext(c)
	r, err := zpd.Listen(ctx, c.GlobalInt("port"))
	if err != nil {
		return err
	}
	defer r.Close()

	p := mpb.NewWithContext(ctx,
		mpb.WithOutput(os.Stderr),
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond),
	)

	var total int64
	bar := p.AddBar(total, mpb.BarStyle("[=>-|"),
		mpb.PrependDecorators(
			decor.CountersKibiByte("% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.Name(" ] "),
			decor.EwmaSpeed(decor.UnitKiB, "% .2f", 60),
		),
	)

	lastTime := time.Now()
	for {
		buf := make([]byte, 32*1024)
		n, err := r.Read(buf)
		total += int64(n)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		_, err = os.Stdout.Write(buf[:n])
		if err != nil {
			return err
		}
		bar.SetTotal(total+2048, false)
		bar.IncrBy(n, time.Since(lastTime))
		lastTime = time.Now()
	}

	bar.SetTotal(total, true)
	p.Wait()
	return nil
}
