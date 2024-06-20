package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/subhammahanty235/artstore-backend/internal/drivers/query"
	"github.com/subhammahanty235/artstore-backend/internal/models"
	"github.com/subhammahanty235/artstore-backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// P1 ---> Save the basic details && images
func (ga *StoreApp) AddNewProductP1(db *mongo.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		type RequestData struct {
			BasicDetails  models.BasicDetails `json:"basicDetails"`
			TransactionId primitive.ObjectID  `json:"transactionId"`
			Images        models.ProductImage `json:"images"`
		}

		var requestData *RequestData
		if err := ctx.ShouldBindJSON(&requestData); err != nil {
			_ = ctx.AbortWithError(http.StatusBadRequest, gin.Error{Err: err})
		}

		dbCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Check the transaction
		var transactionData *models.ShopRequestTransactions
		transactionCollection := query.RequestTransactions(*db)

		findTransactionErr := transactionCollection.FindOne(dbCtx, bson.D{{
			Key:   "transactionId",
			Value: requestData.TransactionId,
		}}).Decode(&transactionData)

		if findTransactionErr == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "Error while finding the transaction details",
				"error":   findTransactionErr.Error(),
				"success": false,
			})
			return
		}

		if transactionData.RequestCreated.Status {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "You have already requested some reviews",
				"error":   "Transaction sent for review",
				"success": false,
			})
			return
		}
		if transactionData.RequestApproved.Status {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "Request already approved",
				"error":   "Transaction sent for review",
				"success": false,
			})
			return
		}
		if transactionData.RequestDeclined.Status {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "Already declined your request",
				"error":   "Transaction sent for review",
				"success": false,
			})
			return
		}

		if requestData.BasicDetails.Name == "" || requestData.BasicDetails.Description == "" || requestData.BasicDetails.Brand == "" {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{
				"message": "Please add all the required details",
				"error":   "Please add all the required details",
				"Success": false,
			})
			return
		}

		collection := query.Product(*db)
		productId := primitive.NewObjectID()
		randomNum, _ := utils.GenerateOtp(3)
		productCode := (strings.ReplaceAll(requestData.BasicDetails.Name, " ", "-")) + randomNum

		_, insertErr := collection.InsertOne(dbCtx, gin.H{
			"basicDetails":  requestData.BasicDetails,
			"_id":           productId,
			"transactionId": requestData.TransactionId,
			"images":        requestData.Images,
			"productCode":   productCode,
		})

		if insertErr != nil {
			var g *query.StoreAppDB
			g.App.ErrorLogger.Fatalf("Cannot add user to database %v", insertErr)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{

			"message": "Added Successfully",
			"success": true,
		})

	}
}

func (ga *StoreApp) AddNewProductP2(db *mongo.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		type RequestData struct {
			ProductPrice  models.ProductPrice  `json:"productPrice"`
			TransactionId primitive.ObjectID   `json:"transactionId"`
			ProductStocks models.ProductStocks `json:"productStocks"`
		}

		var requestData *RequestData
		if err := ctx.ShouldBindJSON(&requestData); err != nil {
			_ = ctx.AbortWithError(http.StatusBadRequest, gin.Error{Err: err})
		}

		dbCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Check the transaction
		var transactionData *models.ShopRequestTransactions
		transactionCollection := query.RequestTransactions(*db)

		findTransactionErr := transactionCollection.FindOne(dbCtx, bson.D{{
			Key:   "transactionId",
			Value: requestData.TransactionId,
		}}).Decode(&transactionData)

		if findTransactionErr == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "Error while finding the transaction details",
				"error":   findTransactionErr.Error(),
				"success": false,
			})
			return
		}

		if transactionData.RequestCreated.Status {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "You have already requested some reviews",
				"error":   "Transaction sent for review",
				"success": false,
			})
			return
		}
		if transactionData.RequestApproved.Status {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "Request already approved",
				"error":   "Transaction sent for review",
				"success": false,
			})
			return
		}
		if transactionData.RequestDeclined.Status {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "Already declined your request",
				"error":   "Transaction sent for review",
				"success": false,
			})
			return
		}

		collection := query.Product(*db)

		_, insertErr := collection.UpdateOne(
			dbCtx,
			bson.D{
				{
					Key:   "transactionId",
					Value: requestData.TransactionId,
				},
			},
			gin.H{
				"price":     requestData.ProductPrice,
				"inventory": requestData.ProductStocks,
			},
		)

		if insertErr != nil {
			var g *query.StoreAppDB
			g.App.ErrorLogger.Fatalf("Cannot add user to database %v", insertErr)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{

			"message": "Added Successfully",
			"success": true,
		})

	}
}

func (ga *StoreApp) FetchProductsbyShopId(db *mongo.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		type ShopCodeInput struct {
			ShopCode string `uri:"shop" binding:"required"`
		}

		var reqData *ShopCodeInput
		if err := ctx.ShouldBindUri(&reqData); err != nil {
			_ = ctx.AbortWithError(http.StatusBadRequest, gin.Error{Err: err})
		}

		dbctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var shopData models.Shop
		shopCollection := query.Shops(*db)

		findShopErr := shopCollection.FindOne(dbctx, bson.D{{
			Key:   "shopCode",
			Value: reqData.ShopCode,
		}}).Decode(&shopData)

		if findShopErr == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message":   "Error while finding the shop details",
				"error":     findShopErr.Error(),
				"success":   false,
				"errorCode": 100001,
			})
			return
		}

		if !shopData.VerificationStatus {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message":   "Shop is not verified yet",
				"error":     findShopErr.Error(),
				"success":   false,
				"errorCode": 100002,
			})
			return
		}

		if shopData.AdminSideData.Status != 1 {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message":   "Shop is not available to deliver",
				"error":     findShopErr.Error(),
				"success":   false,
				"errorCode": 100082,
			})
			return
		}

		shopDataForResponse := gin.H{
			"shopCode":   shopData.ShopCode,
			"images":     shopData.Images,
			"shopName":   shopData.ShopName,
			"address":    shopData.Address,
			"categories": shopData.Categories,
			"telephone":  shopData.Telephone,
			"website":    shopData.Website,
		}

		var productsData []bson.M
		productCollection := query.Product(*db)

		projection := bson.M{
			"basicDetails": 1,
			"images":       1,
			"price":        1,
			"rating":       1,
			"productCode":  1,
		}

		cursor, err := productCollection.Find(dbctx, bson.M{"shopCode": shopData.ShopCode, "adminSideData.status": 1}, options.Find().SetProjection(projection))
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "Failed to fetch",
				"success": false,
			})
			return
		}

		defer cursor.Close(dbctx)
		for cursor.Next(dbctx) {
			var product bson.M
			if err := cursor.Decode(&product); err != nil {
				ctx.JSON(500, gin.H{"error": "Failed to decode shop"})
				return
			}
			productsData = append(productsData, product)
		}
		if err := cursor.Err(); err != nil {
			ctx.JSON(500, gin.H{"error": "Cursor error"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message":      "products Fetched Successfully",
			"success":      true,
			"products":     productsData,
			"productCount": len(productsData),
			"shop":         shopDataForResponse,
		})
	}
}
