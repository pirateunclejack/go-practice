package product

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pirateunclejack/go-practice/dynamodb-go-crud/internal/controllers/product"
	EntityProduct "github.com/pirateunclejack/go-practice/dynamodb-go-crud/internal/entities/product"
	"github.com/pirateunclejack/go-practice/dynamodb-go-crud/internal/handlers"
	"github.com/pirateunclejack/go-practice/dynamodb-go-crud/internal/repository/adapter"
	Rules "github.com/pirateunclejack/go-practice/dynamodb-go-crud/internal/rules"
	RulesProduct "github.com/pirateunclejack/go-practice/dynamodb-go-crud/internal/rules/product"
	HttpStatus "github.com/pirateunclejack/go-practice/dynamodb-go-crud/utils/http"
)

type Handler struct {
    handlers.Interface
    Controller product.Interface
    Rules Rules.Interface
}

func NewHandler(repository adapter.Interface) handlers.Interface {
    return &Handler{
        Controller: product.NewContoller(repository),
        Rules: RulesProduct.NewRules(),
    }
}


func (h *Handler) getOne(w http.ResponseWriter, r *http.Request)  {
    ID, err := uuid.Parse(chi.URLParam(r, "ID"))
    if err != nil {
        HttpStatus.StatusBadRequest(w, r, errors.New("ID is not a valid uuid"))
        return
    }

    response, err := h.Controller.ListOne(ID)
    if err != nil {
        HttpStatus.StatusInternalServerError(w, r, err)
        return
    }

    HttpStatus.StatusOK(w, r, response)
}

func (h *Handler) getAll(w http.ResponseWriter, r *http.Request)  {
    response, err := h.Controller.ListAll()
    if err != nil {
        HttpStatus.StatusInternalServerError(w, r, err)
        return
    }
    HttpStatus.StatusOK(w, r, response)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request)  {
    if chi.URLParam(r, "ID") != "" {
        h.getOne(w, r)
    } else {
        h.getAll(w, r)
    }
}

func setDefaultValues(product *EntityProduct.Product, ID uuid.UUID)  {
    product.UpdatedAt = time.Now()
    if ID == uuid.Nil {
        product.ID = uuid.New()
        product.CreatedAt = time.Now()
    } else {
        product.ID = ID
    }
}


func (h *Handler) getBodyAndValidate(r *http.Request, ID uuid.UUID) (
    *EntityProduct.Product, error) {
        productBody := &EntityProduct.Product{}
        body, err := h.Rules.ConvertIoReaderToStruct(r.Body, productBody)
        if err != nil {
            return &EntityProduct.Product{}, errors.New("body is required")
        }

        productParsed, err := EntityProduct.InterfaceToModel(body)
        if err != nil {
            return &EntityProduct.Product{}, errors.New(
                "error on converting body to model")
        }
        setDefaultValues(productParsed, ID)
        return productParsed, h.Rules.Validate(productParsed)

}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request)  {
    productBody, err := h.getBodyAndValidate(r, uuid.Nil)
    if err != nil {
        HttpStatus.StatusBadRequest(w, r, err)
    }

    ID, err := h.Controller.Create(productBody)
    if err != nil {
        HttpStatus.StatusInternalServerError(w, r, err)
        return
    }

    HttpStatus.StatusOK(w, r, map[string]interface{}{"id": ID.String()})
}

func (h *Handler) Put(w http.ResponseWriter, r *http.Request)  {
    ID, err := uuid.Parse(chi.URLParam(r, "ID"))
    if err != nil {
        HttpStatus.StatusBadRequest(w, r, errors.New("ID is not a valid uuid"))
        return
    }

    productBody, err := h.getBodyAndValidate(r, ID)
    if err != nil {
        HttpStatus.StatusBadRequest(w, r, err)
        return
    }

    if err := h.Controller.Update(ID, productBody); err != nil {
        HttpStatus.StatusInternalServerError(w, r, err)
        return
    }

    HttpStatus.StatusNoContent(w, r)

}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request)  {
    ID, err := uuid.Parse(chi.URLParam(r, "ID"))
    if err != nil {
        HttpStatus.StatusBadRequest(w, r, errors.New("ID is not a valid uuid"))
        return
    }

    if err := h.Controller.Remove(ID); err != nil {
        HttpStatus.StatusInternalServerError(w, r, err)
        return
    }
    
    HttpStatus.StatusNoContent(w, r)
}

func (h *Handler) Options(w http.ResponseWriter, r *http.Request)  {
    HttpStatus.StatusNoContent(w, r)
}

