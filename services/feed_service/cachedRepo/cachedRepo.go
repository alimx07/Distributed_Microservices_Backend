package cachedrepo

import "github.com/alimx07/Distributed_Microservices_Backend/servicesfeed_service/models"

type Cache interface {
	Set(models.FeedItem) error
	Get(models.Cursor) ([]models.FeedItem, string, error)
	Close() error
}
