package constants

// ServiceName The name of this module/service
const ServiceName = "restaurants"

// GRPC Service Names
const (
	RestaurantsServiceName = "RESTAURANTS"
)

// Dependency Injection Keys
const (
	RegistryKey                 = "registry"
	DomainDispatcherKey         = "domainDispatcher"
	DatabaseTransactionKey      = "tx"
	MessagePublisherKey         = "messagePublisher"
	MessageSubscriberKey        = "messageSubscriber"
	EventPublisherKey           = "eventPublisher"
	CommandPublisherKey         = "commandPublisher"
	ReplyPublisherKey           = "replyPublisher"
	AggregateStoreKey           = "aggregateStore"
	SagaStoreKey                = "sagaStore"
	InboxRestaurantKey          = "inboxRestaurant"
	ApplicationKey              = "app"
	DomainEventHandlersKey      = "domainEventHandlers"
	IntegrationEventHandlersKey = "integrationEventHandlers"
	CommandHandlersKey          = "commandHandlers"
	ReplyHandlersKey            = "replyHandlers"

	MallHandlersKey = "mallHandlers"

	RestaurantsRepoKey = "restaurantsRepo"
	MallRepoKey        = "mallRepo"
	//StoresRepoKey   = "storesRepo"
	//ProductsRepoKey = "productsRepo"
	//CatalogRepoKey  = "catalogRepo"
)
