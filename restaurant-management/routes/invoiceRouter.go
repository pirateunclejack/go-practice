package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/pirateunclejack/go-practice/restaurant-management/controllers"
)

func InvoiceRoutes(incomingRoutes *gin.Engine)  {
    incomingRoutes.GET("/invoices", controller.GetInvoices())
    incomingRoutes.GET("/invoice/:invoice_id", controller.GetInvoice())
    incomingRoutes.POST("/invoice", controller.CreateInvoice())
    incomingRoutes.PATCH("/invoice/:invoice_id", controller.UpdateInvoice())
}
