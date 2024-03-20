package database

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/pirateunclejack/go-practice/ecommerce-yt/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrCantFindProduct = errors.New("can't find the product")
	ErrCantDecodeProducts = errors.New("can't find the product")
	ErrUserIdIsNotValid = errors.New("this user is not valid")
	ErrCantUpdateUser = errors.New("can't add this product to the cart")
	ErrCantRemoveItemCart = errors.New("can't remove this item from the cart")
	ErrCantGetItem = errors.New("was unable to get the item from the cart")
	ErrCantBuyCartItem = errors.New("can't update the purchase")
)

func AddProductToCart(
	ctx context.Context,
	prodCollection, userCollection *mongo.Collection,
	productID primitive.ObjectID,
	userID string,
	) error {
		searchfromdb, err := prodCollection.Find(ctx, bson.M{
			"_id": productID,
		})
		if err != nil {
			log.Println(err)
			return ErrCantFindProduct
		}

		var productCart []models.ProductUser
		err = searchfromdb.All(ctx, &productCart)
		if err!= nil {
			log.Println(err)
			return ErrCantDecodeProducts
		}

		id, err := primitive.ObjectIDFromHex(userID)
		log.Printf("provided user object id: %v", id)
		if err != nil {
			log.Printf("failed to get objectid from hex: %v", err)
			return ErrUserIdIsNotValid
		}
		var user models.User
		err = userCollection.FindOne(ctx, bson.M{
			"_id": id,
		}).Decode(&user)
		if err != nil {
			log.Printf("failed to find user with provided id %v,  error: %v", userID, err)
			return ErrUserIdIsNotValid
		}

		filter := bson.D{primitive.E{
			Key: "_id",
			Value: id,
		}}
		update := bson.D{{
			Key: "$push",
			Value: bson.D{{
				Key: "usercart",
				Value: bson.D{{
					Key: "$each",
					Value: productCart,
				}},
			}},
		}}

		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			return ErrCantUpdateUser
		}
		return nil

}

func RemoveCartItem(
	ctx context.Context,
	prodCollection, userCollection *mongo.Collection,
	productID primitive.ObjectID,
	userID string,
	) error {
		id, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			log.Println(err)
			return ErrUserIdIsNotValid
		}

		filter := bson.D{primitive.E{Key: "_id", Value: id}}
		update := bson.M{"$pull": bson.M{"usercart": bson.M{"_id": productID}}}
		_, err = userCollection.UpdateMany(ctx, filter, update) //
		if err != nil {
			return ErrCantRemoveItemCart
		}

		return nil
}

func BuyItemFromCart(
	ctx context.Context,
	userCollection *mongo.Collection,
	userID string,
) error {
	// fetch the cart from the user
	// find the cart total
	// create a order with items
	// added order to the user collection
	// added items in the cart to order list
	// empty up the cart

	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}

	var getcartitems models.User
	var ordercart models.Order

	ordercart.Order_ID = primitive.NewObjectID()
	ordercart.Ordered_At = time.Now()
	ordercart.Order_Cart = make([]models.ProductUser, 0)
	ordercart.Payment_Method.COD = true

	unwind := bson.D{{Key: "$unwind", Value: bson.D{
		primitive.E{Key: "path", Value: "$usercart"},
	}}}
	grouping := bson.D{{Key: "$group", Value: bson.D{
		primitive.E{Key: "_id", Value: "$_id"},
		{Key: "total", Value: bson.D{
			primitive.E{Key: "$sum", Value: "$usercart.price"},
		}},
	}}}

	currentresults, err := userCollection.Aggregate(
		ctx, mongo.Pipeline{unwind, grouping})
	ctx.Done()
	if err != nil {
		panic(err)
	}

	var getusercart []bson.M
	if err := currentresults.All(ctx, &getusercart); err != nil {
		panic(err)
	}

	var total_price int32

	for _, user_item := range getusercart {
		price := user_item["total"]
		total_price = price.(int32)
	}

	ordercart.Price = int(total_price)

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{
		Key: "$push", 
		Value: bson.D{primitive.E{
			Key: "orders",
			Value: ordercart,
		}},
	}}
	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Println(err)
	}

	err = userCollection.FindOne(ctx, bson.D{
		primitive.E{
			Key: "_id",
			Value: id,
		},
	}).Decode(&getcartitems)
	if err != nil {
		log.Println(err)
	}

	filter2 := bson.D{primitive.E{Key: "_id", Value: id}}
	update2 := bson.M{"$push": bson.M{
		"orders.$[].order_list": bson.M{"$each": getcartitems.UserCart}}}
	userCollection.UpdateOne(ctx, filter2, update2)

	if err != nil {
		log.Println(err)
	}

	usercart_empty := make([]models.ProductUser, 0)
	filter3 := bson.D{primitive.E{Key: "_id", Value: id}}
	update3 := bson.D{{
		Key: "$set",
		Value: bson.D{primitive.E{
			Key: "usercart",
			Value: usercart_empty,
		},
	}}}
	_, err = userCollection.UpdateOne(ctx, filter3, update3)
	if err != nil {
		return ErrCantBuyCartItem
	}

	return nil
}

func InstantBuyer(
	ctx context.Context,
	prodCollection, userCollection *mongo.Collection,
	productID primitive.ObjectID,
	userID string,
	) error {
		id, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			log.Println(err)
			return ErrUserIdIsNotValid
		}

		var product_details models.ProductUser
		var orders_detail models.Order

		orders_detail.Order_ID = primitive.NewObjectID()
		orders_detail.Ordered_At = time.Now()
		orders_detail.Order_Cart = make([]models.ProductUser, 0)
		orders_detail.Payment_Method.COD = true
		err = prodCollection.FindOne(ctx, bson.D{primitive.E{
			Key: "_id",
			Value: productID,
		}}).Decode(&product_details)
		if err != nil {
			log.Println(err)
		}

		orders_detail.Price = product_details.Price

		filter := bson.D{primitive.E{Key: "_id", Value: id}}
		update := bson.D{{
			Key: "$push", 
			Value: bson.D{primitive.E{
				Key: "orders",
				Value: orders_detail,
			}},
		}}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Println(err)
		}

		filter2 := bson.D{primitive.E{Key: "_id", Value: id}}
		update2 := bson.M{"$push": bson.M{"orders.$[].order_list":product_details}}
		_, err = userCollection.UpdateOne(ctx, filter2, update2)
		if err != nil {
			log.Println(err)
		}

		return nil
}