package mongodb

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

func TestMongoDriver_CreateIndexes(t *testing.T) {
	uri := "mongodb://admin:xxxx@xxx.xx.xx.xxxx:27017"
	m, err := NewMongoDriver(uri)
	if err != nil {
		t.Error(err)
		return
	}
	// 设置bson_attach 字段 ,支持索引 优先级,唯一索引 复合索引
	type User struct {
		ID       primitive.ObjectID `bson:"_id" json:"id"`
		UserName string             `bson:"name" json:"name" bson_attach:"index:index_user,priority:1,unique"`
		Age      int                `bson:"age" json:"age"`
		Email    string             `bson:"email" json:"email" bson_attach:"index:index_user,priority:2" `
	}
	m.SetDatabase("cmdb")
	userTable := "test_user"
	//newUser := User{
	//	ID:       primitive.NewObjectID(),
	//	UserName: "test123",
	//	Age:      18,
	//}
	//_, err = m.client.Database(m.dbName).Collection(userTable).InsertOne(m.ctx, newUser)
	//if err != nil {
	//	t.Error(err)
	//	return
	//}

	models := []map[string]interface{}{
		{
			"tableName": userTable,
			"model":     User{},
		},
	}
	m.CreateIndexes(models, true)
}
