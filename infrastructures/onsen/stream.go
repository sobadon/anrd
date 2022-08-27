package onsen

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sobadon/anrd/domain/model/program"
	"github.com/sobadon/anrd/domain/model/recorder"
	"github.com/sobadon/anrd/internal/errutil"
	"github.com/sobadon/anrd/internal/fileutil"
)

func (c *client) Rec(ctx context.Context, config recorder.Config, targetPgram program.Program) error {
	file := buildArchiveFilePath(config.ArchiveDir, targetPgram)
	err := fileutil.MkdirAllIfNotExist(filepath.Dir(file))
	if err != nil {
		return errors.Wrap(errutil.ErrInternal, err.Error())
	}

	cmd := exec.Command("ffmpeg",
		"-y",
		"-loglevel", "warning", // とりあえず決め打ち
		"-headers", "Referer: https://www.onsen.ag/'$'\r\n", // リファラがなければ 403
		"-i", targetPgram.PlaylistURL,
		"-vcodec", "copy",
		"-acodec", "copy",
		file,
	)

	log.Println(cmd.String())

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return errors.Wrap(errutil.ErrFfmpeg, err.Error())
	}

	err = cmd.Wait()
	if err != nil {
		return errors.Wrap(errutil.ErrFfmpeg, err.Error())
	}

	return nil
}

func buildArchiveFilePath(basePath string, pgram program.Program) string {
	return filepath.Join(
		basePath,
		"onsen",
		fileutil.SanitizeReplaceName(pgram.Title),
		fmt.Sprintf("%s_%s_%s.ts", pgram.Start.Format("2006-01-02"), fileutil.SanitizeReplaceName(pgram.Title), fileutil.SanitizeReplaceName(pgram.Episode)))
}
