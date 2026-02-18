package users

import (
	"context"
	"errors"
	"time"

	"github.com/pat3icki/pennychoice/pkg/sflake"
	"github.com/pat3icki/pennychoice/pkg/utils"
)

func (s *Service) CreateRequestKey(ctx context.Context, req *RequestKey) error {
	if req == nil {
		return errors.New("parameter(s) cannot be nil")
	}
	var (
		err     error
		key     int64
		purpose string
		tll     time.Time
	)
	time.Sleep(500 * time.Millisecond)
	key, err = sflake.GenerateID(sflake.DefaultEpoch, REQUEST_KEY_SFLAKE_NODE, &sflake.State{})
	if err != nil {
		return err
	}
	if (req.Period.Unix() - time.Now().Unix()) < 0 {
		return errors.New("req.period cannot be less than current timestamp")
	}
	tll = req.Period

	if (len(req.Purpose) == 0) || req.Purpose == "" {
		return errors.New("req.purpose cannot be empty")
	}
	purpose = req.Purpose

	err = s.Redis.Set(utils.Convert.IntBytes(key), []byte(purpose), tll)
	if err != nil {
		return err
	}
	req.ID = key
	return nil

}
