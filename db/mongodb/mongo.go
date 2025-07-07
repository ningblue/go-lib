package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

/*
file: 用于初始化mongo数据库连接
*/

func BuildURI(address string, port int, user, password, database string) string {
	if !strings.Contains(address, ":") && port > 0 {
		address = address + ":" + strconv.Itoa(port)
	}

	uri := fmt.Sprintf("mongodb://%s:%s@%s/%s", user, password, address, database)
	return uri
}

type MongoDriver struct {
	client     *mongo.Client
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
	dbName     string
	collection string
	instanceID string
}

func NewMongoOptions(uri string) (*options.ClientOptions, error) {
	conn, err := connstring.ParseAndValidate(uri)
	if err != nil {
		return nil, err
	}

	credential := options.Credential{
		Username: conn.Username,
		Password: conn.Password,
	}
	if conn.HasAuthParameters() {
		credential.AuthMechanism = conn.AuthMechanism
	}
	clientOptions := options.Client().ApplyURI(conn.String()).SetAuth(credential)
	return clientOptions, nil
}
func NewMongoDriver(uri string) (*MongoDriver, error) {
	ctx := context.Background()
	clientOptions, err := NewMongoOptions(uri)
	if err != nil {
		return nil, err
	}
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}
	newMongo := &MongoDriver{client: client, ctx: ctx, cancel: nil}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	dbName := u.Path[1:]
	newMongo.SetDatabase(dbName)
	return newMongo, nil
}

// 做一个默认的连接
const defaultMongoDBURI = "mongodb://localhost:27017"

func defaultMongoDriver() (*MongoDriver, error) {
	return NewMongoDriver(defaultMongoDBURI)
}

// CheckDatabase 检查
func (m *MongoDriver) CheckDatabase() error {
	return m.ping()
}

// ping 检查数据库连接是否正常
func (m *MongoDriver) ping() error {
	return m.client.Ping(m.ctx, nil)
}

// Close 关闭数据库连接
func (m *MongoDriver) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.client.Disconnect(m.ctx); err != nil {
		return err
	}
	m.cancel()
	return nil
}

// GetTable 获取表 get  db collection
func (m *MongoDriver) GetTable(collName string) Table {
	col := Collection{}
	col.collName = collName
	col.MongoDriver = m
	return &col
}

func (m *MongoDriver) SetDatabase(dbName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dbName = dbName
}

// GetDatabase .
func (m *MongoDriver) GetDatabase() string {
	return m.dbName
}

// getCurrentDatabase 获取当前数据库
func (m *MongoDriver) getCurrentDatabase() (string, error) {
	if m.dbName != "" {
		return m.dbName, nil
	}
	dbs, err := m.client.ListDatabaseNames(m.ctx, bson.D{})
	if err != nil {
		return "", fmt.Errorf("failed to select default database: %w", err)
	}
	if len(dbs) < 1 {
		return "", fmt.Errorf("no databases found")
	}
	m.dbName = dbs[0]
	return m.dbName, nil
}

// CreateTable 创建表  create collection
func (m *MongoDriver) CreateTable(collName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.client.Database(m.dbName).RunCommand(m.ctx, map[string]interface{}{"create": collName}).Err()
}

// DeleteTable 删除表  drop collection
func (m *MongoDriver) DeleteTable(collName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.client.Database(m.dbName).Collection(collName).Drop(m.ctx)
}

// RenameTable  更新表名  rename collection
func (m *MongoDriver) RenameTable(ctx context.Context, preName, newName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.client.Database(m.dbName).RunCommand(ctx, map[string]interface{}{"renameCollection": preName, "to": newName}).Err()
}

// HasTable 判断表是否存在  check collection exist
func (m *MongoDriver) HasTable(collName string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	collections, err := m.client.Database(m.dbName).ListCollectionNames(m.ctx, bson.D{})
	if err != nil {
		return false, err
	}
	for _, collection := range collections {
		if collection == collName {
			return true, nil
		}
	}
	return false, nil
}

// ListTables 获取所有表名 list all collections
func (m *MongoDriver) ListTables(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.client.Database(m.dbName).ListCollectionNames(ctx, bson.D{})
}

func getCollectionOption(ctx context.Context) *options.CollectionOptions {
	var opt *options.CollectionOptions
	switch GetDBReadPreference(ctx) {

	case NilMode:

	case PrimaryMode:
		opt = &options.CollectionOptions{
			ReadPreference: readpref.Primary(),
		}
	case PrimaryPreferredMode:
		opt = &options.CollectionOptions{
			ReadPreference: readpref.PrimaryPreferred(readpref.WithMaxStaleness(maxStalenessSeconds)),
		}
	case SecondaryMode:
		opt = &options.CollectionOptions{
			ReadPreference: readpref.Secondary(readpref.WithMaxStaleness(maxStalenessSeconds)),
		}
	case SecondaryPreferredMode:
		opt = &options.CollectionOptions{
			ReadPreference: readpref.SecondaryPreferred(readpref.WithMaxStaleness(maxStalenessSeconds)),
		}
	case NearestMode:
		opt = &options.CollectionOptions{
			ReadPreference: readpref.Nearest(readpref.WithMaxStaleness(maxStalenessSeconds)),
		}
	}

	return opt
}
