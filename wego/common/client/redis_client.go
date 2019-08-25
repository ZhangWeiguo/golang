package client

import (
	redis "github.com/go-redis/redis/v7"
	"time"
)

type RedisClient struct {
	Addr         string
	Pass         string
	DB           int
	PoolSize     int
	WriteTimeOut time.Duration
	ReadTimeOut  time.Duration
	DialTimeOut  time.Duration
	NetWork      string
	client       *redis.Client
}

func (rc *RedisClient) Init() error {
	rc.client = redis.NewClient(&redis.Options{
		Addr:         rc.Addr,
		Password:     rc.Pass,
		DB:           rc.DB,
		PoolSize:     rc.PoolSize,
		WriteTimeout: rc.WriteTimeOut,
		ReadTimeout:  rc.ReadTimeOut,
		DialTimeout:  rc.DialTimeOut,
		Network:      rc.NetWork,
	})
	_, err := rc.client.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

// ========== String
func (rc *RedisClient) Get(key string) (string, error) {
	return rc.client.Get(key).Result()
}

// 设置字符串 过期时间<=0则永不过期
func (rc *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return rc.client.Set(key, value, expiration).Err()
}

// 键不存在时才设置
func (rc *RedisClient) SetNX(key string, value interface{}, expiration time.Duration) error {
	return rc.client.SetNX(key, value, expiration).Err()
}

// 键存在时才设置
func (rc *RedisClient) SetXX(key string, value interface{}, expiration time.Duration) error {
	return rc.client.SetXX(key, value, expiration).Err()
}

// ========== HashMap
func (rc *RedisClient) GetHash(key string) (map[string]string, error) {
	return rc.client.HGetAll(key).Result()
}

func (rc *RedisClient) GetHashKey(key string, field string) (string, error) {
	return rc.client.HGet(key, field).Result()
}

// 设置HashMap
func (rc *RedisClient) SetHash(key string, value map[string]interface{}) error {
	return rc.client.HMSet(key, value).Err()
}

// 设置HashMap的一个key
func (rc *RedisClient) SetHashKey(key string, field string, value interface{}) error {
	result := rc.client.HSet(key, field, value)
	return result.Err()
}

// 设置HashMap的一个field 当且仅当field不存在
func (rc *RedisClient) SetHashKeyNX(key string, field string, value interface{}) error {
	return rc.client.HSetNX(key, field, value).Err()
}

// ========== List
func (rc *RedisClient) GetList(key string, start int64, end int64) ([]string, error) {
	return rc.client.LRange(key, start, end).Result()
}

func (rc *RedisClient) GetListPop(key string, start int64, end int64) (string, error) {
	return rc.client.LPop(key).Result()
}

func (rc *RedisClient) GetListLen(key string) (int64, error) {
	return rc.client.LLen(key).Result()
}

func (rc *RedisClient) PushList(key string, value interface{}) error {
	return rc.client.LPush(key, value).Err()
}

func (rc *RedisClient) SetList(key string, index int64, value interface{}) error {
	return rc.client.LSet(key, index, value).Err()
}

// 插入到pivot后面
func (rc *RedisClient) InsertListAfter(key string, pivot, value interface{}) error {
	return rc.client.LInsertAfter(key, pivot, value).Err()
}

// 插入到pivot前面
func (rc *RedisClient) InsertListBefore(key string, pivot, value interface{}) error {
	return rc.client.LInsertBefore(key, pivot, value).Err()
}

// ========== Set
func (rc *RedisClient) GetSet(key string) ([]string, error) {
	return rc.client.SMembers(key).Result()
}

func (rc *RedisClient) GetSets(keys ...string) ([]string, error) {
	return rc.client.SUnion(keys...).Result()
}

func (rc *RedisClient) GetSetPop(key string) (string, error) {
	return rc.client.SPop(key).Result()
}

func (rc *RedisClient) GetSetLen(key string) (int64, error) {
	return rc.client.SCard(key).Result()
}

func (rc *RedisClient) GetSetIsIn(key string, value interface{}) (bool, error) {
	return rc.client.SIsMember(key, value).Result()
}

func (rc *RedisClient) SetSetAdd(key string, values ...interface{}) error {
	return rc.client.SAdd(key, values).Err()
}

// ========== SortedSet
func (rc *RedisClient) GetZSet(key string, start, end int64) ([]string, error) {
	return rc.client.ZRange(key, start, end).Result()
}

func (rc *RedisClient) GetZSetLen(key string) (int64, error) {
	return rc.client.ZCard(key).Result()
}

func (rc *RedisClient) SetZSetAdd(key string, values ...*redis.Z) error {
	return rc.client.ZAdd(key, values...).Err()
}

// ========== 过期时间
// 秒级别的过期时间
func (rc *RedisClient) SetExpire(key string, expiration time.Duration) error {
	return rc.client.Expire(key, expiration).Err()
}

// 秒级别的过期时间点
func (rc *RedisClient) SetExpiredAt(key string, at time.Time) error {
	return rc.client.ExpireAt(key, at).Err()
}

// 毫秒级别的过期时间
func (rc *RedisClient) SetPExpire(key string, expiration time.Duration) error {
	return rc.client.PExpire(key, expiration).Err()
}

// 毫秒级别的过期时间点
func (rc *RedisClient) SetPExpiredAt(key string, at time.Time) error {
	return rc.client.PExpireAt(key, at).Err()
}

// ========== 删除Key
func (rc *RedisClient) Delete(keys ...string) (int64, error) {
	return rc.client.Del(keys...).Result()
}

// ========== 订阅发布
func (rc *RedisClient) Publish(channel string, message interface{}) error {
	return rc.client.Publish(channel, message).Err()
}

func (rc *RedisClient) SubScribe(channel string) *redis.PubSub {
	return rc.client.Subscribe(channel)
}

func (rc *RedisClient) GetClient() *redis.Client {
	return rc.client
}

func (rc *RedisClient) Close() error {
	return rc.client.Close()
}
