package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrLockNotAcquired = errors.New("seat is already locked by another user")
var ErrLockNotOwned = errors.New("lock is not owned by this token")

type LockService struct {
	redisClient *redis.Client
	lockTTL     time.Duration
}

func NewLockService(redisClient *redis.Client, lockTTL time.Duration) *LockService {
	return &LockService{
		redisClient: redisClient,
		lockTTL:     lockTTL,
	}
}

func seatLockKey(showtimeID, seatID string) string {
	return fmt.Sprintf("seat_lock:%s:%s", showtimeID, seatID)
}

func (s *LockService) AcquireLock(ctx context.Context, showtimeID, seatID, lockToken string) error {
	key := seatLockKey(showtimeID, seatID)

	acquired, err := s.redisClient.SetNX(ctx, key, lockToken, s.lockTTL).Result()
	if err != nil {

		return fmt.Errorf("redis error while acquiring lock: %w", err)
	}

	if !acquired {
		return ErrLockNotAcquired
	}

	return nil
}

var releaseLockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end
`)

func (s *LockService) ReleaseLock(ctx context.Context, showtimeID, seatID, lockToken string) error {
	key := seatLockKey(showtimeID, seatID)

	result, err := releaseLockScript.Run(ctx, s.redisClient, []string{key}, lockToken).Result()
	if err != nil {
		return fmt.Errorf("redis error while releasing lock: %w", err)
	}

	deleted, ok := result.(int64)
	if !ok || deleted == 0 {
		return ErrLockNotOwned
	}

	return nil
}

func (s *LockService) GetLockOwner(ctx context.Context, showtimeID, seatID string) (string, error) {
	key := seatLockKey(showtimeID, seatID)
	val, err := s.redisClient.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {

		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("redis error while checking lock: %w", err)
	}
	return val, nil
}
