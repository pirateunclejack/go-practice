package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pirateunclejack/go-practice/restaurant-management/database"
	"github.com/pirateunclejack/go-practice/restaurant-management/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderItemPack struct {
    Tabld_id    *string
    Order_items []models.OrderItem
}

var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client, "orderItem")

func GetOrderItems() gin.HandlerFunc {
    return func(c *gin.Context) {
        var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        result, err := orderItemCollection.Find(ctx, bson.M{})
        if err != nil {
            c.JSON(
                http.StatusInternalServerError,
                gin.H{
                    "error": "failed to get order items: " + err.Error(),
                },
            )
            return
        }

        var allOrderItems []bson.M
        if err = result.All(ctx, &allOrderItems); err != nil {
            log.Fatal("failed to save allOrderItems: ", err.Error())
            return
        }

        c.JSON(http.StatusOK, allOrderItems)
    }
}

func ItemsByOrder(id string) (OrderItems []primitive.M, err error) {
    var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    matchStage := bson.D{
        {Key: "$match", Value: bson.D{{Key: "order_id", Value: id}}}}
    lookupStage := bson.D{{Key: "lookup", Value: bson.D{
        {Key: "from", Value: "food"},
        {Key: "localField", Value: "food_id"},
        {Key: "foreignField", Value: "food_id"},
        {Key: "as", Value: "food"},
    }}}
    unwindStage := bson.D{
        {Key: "unwind", Value: bson.D{
            {Key: "path", Value: "$food"},
            {Key: "preserveNullAndEmptyArrays", Value: true},
        }},
    }
    lookupOrderStage := bson.D{
        {Key: "$lookup", Value: bson.D{
            {Key: "from", Value: "order"}, 
            {Key: "localField", Value: "order_id"}, 
            {Key: "foreignField", Value: "order_id"}, 
            {Key: "as", Value: "order"}},
        },
    }
	unwindOrderStage := bson.D{{
        Key: "$unwind", Value: bson.D{
            {Key: "path", Value: "$order"}, 
            {Key: "preserveNullAndEmptyArrays", Value: true},
        }},
    }

	lookupTableStage := bson.D{
        {Key: "$lookup", Value: bson.D{
            {Key: "from", Value:"table"}, 
            {Key: "localField", Value:"order.table_id"}, 
            {Key: "foreignField", Value:"table_id"}, 
            {Key: "as", Value:"table"},
        }},
    }
	unwindTableStage := bson.D{
        {Key: "$unwind", Value: bson.D{
            {Key: "path", Value: "$table"}, 
            {Key: "preserveNullAndEmptyArrays", Value: true},
        }},
    }

	projectStage := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "id", Value: 0},
			{Key: "amount", Value: "$food.price"},
			{Key: "total_count", Value: 1},
			{Key: "food_name", Value: "$food.name"},
			{Key: "food_image", Value: "$food.food_image"},
			{Key: "table_number", Value: "$table.table_number"},
			{Key: "table_id", Value: "$table.table_id"},
			{Key: "order_id", Value: "$order.order_id"},
			{Key: "price", Value: "$food.price"},
			{Key: "quantity", Value: 1},
		}},
    }

	groupStage := bson.D{{Key: "$group", Value: bson.D{
        {Key: "_id", Value: bson.D{
            {Key: "order_id", Value: "$order_id"}, 
            {Key: "table_id", Value: "$table_id"}, 
            {Key: "table_number", Value: "$table_number"},
            }}, 
        {Key: "payment_due", Value: bson.D{
            {Key: "$sum Value:", Value: "$amount"}}}, 
            {Key: "total_count", Value: bson.D{
                {Key: "$sum", Value: 1}}}, 
                {Key: "order_items", Value: bson.D{
                    {Key: "$push", Value: "$$ROOT"}}}}}}

	projectStage2 := bson.D{
		{Key: "$project", Value: bson.D{

			{Key: "id", Value: 0},
			{Key: "payment_due", Value: 1},
			{Key: "total_count", Value: 1},
			{Key: "table_number", Value: "$_id.table_number"},
			{Key: "order_items", Value: 1},
		}}}

	result, err := orderItemCollection.Aggregate(
        ctx, 
        mongo.Pipeline{
		matchStage,
		lookupStage,
		unwindStage,
		lookupOrderStage,
		unwindOrderStage,
		lookupTableStage,
		unwindTableStage,
		projectStage,
		groupStage,
		projectStage2})

	if err != nil {
		panic(err)
	}

	if err = result.All(ctx, &OrderItems); err != nil {
		panic(err)
	}

	defer cancel()

	return OrderItems, err

}

func GetOrderItemsByOrder() gin.HandlerFunc {
    return func(c *gin.Context) {
        orderId := c.Param("order_id")

        allOrderItems, err := ItemsByOrder(orderId)
        if err != nil {
            c.JSON(
                http.StatusInternalServerError,
                gin.H{
                    "error": "failed to get order items by order: " + err.Error(),
                },
            )
            return
        }

        c.JSON(http.StatusOK, allOrderItems)
    }
}

func GetOrderItem() gin.HandlerFunc {
    return func(c *gin.Context) {
        var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        orderItemId := c.Param("orderItem_id")
        var orderItem models.OrderItem

        err := orderItemCollection.FindOne(ctx, bson.M{
            "order_item_id": orderItemId,
        }).Decode(&orderItem)
        if err != nil {
            c.JSON(
                http.StatusInternalServerError,
                gin.H{
                    "error": "failed to save orderItem: " + err.Error(),
                },
            )
            return
        }

        c.JSON(http.StatusOK, orderItem)
    }
}

func CreateOrderItem() gin.HandlerFunc {
    return func(c *gin.Context) {
        var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        var orderItemPack OrderItemPack
        var order models.Order

        if err := c.BindJSON(&orderItemPack); err != nil {
            c.JSON(
                http.StatusBadRequest,
                gin.H{
                    "error": "failed to parse orderItemPack: " + err.Error(),
                },
            )
            return
        }

        order.Order_Date, _ = time.Parse(
            time.RFC3339,
            time.Now().Format(time.RFC3339),
        )
        orderItemsToBeInserted := []interface{}{}
        order.Table_id = orderItemPack.Tabld_id
        order_id := OrderItemOrderCreator(order)

        for _, orderItem := range orderItemPack.Order_items {
            orderItem.Order_id = order_id
            validationErr := validate.Struct(orderItem)
            if validationErr != nil {
                c.JSON(
                    http.StatusBadRequest,
                    gin.H{"error": "failed to valiedate orderItem: " + validationErr.Error()},
                )
                return
            }

            orderItem.ID = primitive.NewObjectID()
            orderItem.Created_at, _ = time.Parse(
                time.RFC3339,
                time.Now().Format(time.RFC3339),
            )
            orderItem.Updated_at, _ = time.Parse(
                time.RFC3339,
                time.Now().Format(time.RFC3339),
            )
            orderItem.Order_item_id = orderItem.ID.Hex()
            var num = toFixed(*orderItem.Unit_price, 2)
            orderItem.Unit_price = &num
            orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)
        }

        insertedOrderItems, err := orderItemCollection.InsertMany(
            ctx,
            orderItemsToBeInserted,
        )
        if err != nil {
            log.Fatal("failed to insert orderItemsToBeInserted: ", err.Error())
            return
        }

        c.JSON(http.StatusOK, insertedOrderItems)

    }
}

func UpdateOrderItem() gin.HandlerFunc {
    return func(c *gin.Context) {
        var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        var orderItem models.OrderItem
        orderItemId := c.Param("orderItem_id")
        filter := bson.M{"order_item_id": orderItemId}

        var updateObj primitive.D
        if orderItem.Unit_price != nil {
            updateObj = append(updateObj, bson.E{
                Key: "unit_price",
                Value: orderItem.Unit_price,
            })
        }
        if orderItem.Quantity != nil {
            updateObj = append(updateObj, bson.E{
                Key: "quantity",
                Value: orderItem.Quantity,
            })
        }
        if orderItem.Food_id != nil {
            updateObj = append(updateObj, bson.E{
                Key: "food_id",
                Value: orderItem.Food_id,
            })
        }

        orderItem.Updated_at, _ = time.Parse(
            time.RFC3339,
            time.Now().Format(time.RFC3339),
        )

        updateObj = append(updateObj, bson.E{
            Key: "updated_at",
            Value: orderItem.Updated_at,
        })

        upsert := true
        opt := options.UpdateOptions{
            Upsert: &upsert,
        }

        result, err := orderItemCollection.UpdateOne(
            ctx,
            filter,
            bson.D{{Key: "$set", Value: updateObj}},
            &opt,
        )
        if err != nil {
            msg := "failed to update order item: " + err.Error()
            c.JSON(
                http.StatusInternalServerError,
                gin.H{"error": msg},
            )
            return
        }

        c.JSON(http.StatusOK, result)
    }
}
