package onsen

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	// https://github.com/rs/zerolog/issues/398
	// log.Level(zerolog.InfoLevel).With().Logger() などとしても
	// 出力されるログに loglevel が含まれない
	cmd.Stdout = log.Ctx(ctx).With().Str("level", zerolog.LevelInfoValue).Logger()
	cmd.Stderr = log.Ctx(ctx).With().Str("level", zerolog.LevelWarnValue).Logger()

	log.Ctx(ctx).Debug().Msgf("ffmpeg start ... (program = %+v)", targetPgram)
	log.Ctx(ctx).Debug().Msgf("ffmpeg command: %s", cmd.String())

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
