package order

import (
	"context"
	"fmt"
	"github.com/lynbklk/tradebot/pkg/notifier"
	"github.com/lynbklk/tradebot/pkg/storage"
	"math"
	"sync"
	"time"

	"github.com/lynbklk/tradebot/pkg/exchange"
	"github.com/lynbklk/tradebot/pkg/model"
	log "github.com/sirupsen/logrus"
)

type Controller struct {
	mtx      sync.Mutex
	ctx      context.Context
	exchange exchange.Exchange
	storage  storage.Storage
	//orderFeed      *Feed
	monitor        Monitor
	notifier       notifier.Notifier
	Results        map[string]*summary
	lastPrice      map[string]float64
	tickerInterval time.Duration
	finish         chan bool
	status         Status
}

func NewController(ctx context.Context, exchange exchange.Exchange, storage storage.Storage,
	monitor Monitor) *Controller {

	return &Controller{
		ctx:      ctx,
		storage:  storage,
		exchange: exchange,
		//orderFeed:      orderFeed,
		monitor:        monitor,
		lastPrice:      make(map[string]float64),
		Results:        make(map[string]*summary),
		tickerInterval: time.Second,
		finish:         make(chan bool),
	}
}

func (c *Controller) SetNotifier(notifier notifier.Notifier) {
	c.notifier = notifier
}

func (c *Controller) OnCandle(candle model.Candle) {
	c.lastPrice[candle.Pair] = candle.Close
}

func (c *Controller) calculateProfit(o *model.Order) (value, percent float64, err error) {
	// get filled orders before the current order
	orders, err := c.storage.Orders(
		storage.WithUpdateAtBeforeOrEqual(o.UpdatedAt),
		storage.WithStatus(model.OrderStatusTypeFilled),
		storage.WithPair(o.Pair),
	)
	if err != nil {
		return 0, 0, err
	}

	quantity := 0.0
	avgPrice := 0.0

	for _, order := range orders {
		// skip current order
		if o.ID == order.ID {
			continue
		}

		// calculate avg price
		if order.Side == model.SideTypeBuy {
			price := order.Price
			if order.Type == model.OrderTypeStopLoss || order.Type == model.OrderTypeStopLossLimit {
				price = *order.Stop
			}
			avgPrice = (order.Quantity*price + avgPrice*quantity) / (order.Quantity + quantity)
			quantity += order.Quantity
		} else {
			quantity = math.Max(quantity-order.Quantity, 0)
		}
	}

	if quantity == 0 {
		return 0, 0, nil
	}

	cost := o.Quantity * avgPrice
	price := o.Price
	if o.Type == model.OrderTypeStopLoss || o.Type == model.OrderTypeStopLossLimit {
		price = *o.Stop
	}
	profitValue := o.Quantity*price - cost
	return profitValue, profitValue / cost, nil
}

func (c *Controller) notify(message string) {
	log.Info(message)
	if c.notifier != nil {
		c.notifier.Notify(message)
	}
}

func (c *Controller) notifyError(err error) {
	log.Error(err)
	if c.notifier != nil {
		c.notifier.OnError(err)
	}
}

func (c *Controller) processTrade(order *model.Order) {
	if order.Status != model.OrderStatusTypeFilled {
		return
	}

	// initializer results map if needed
	if _, ok := c.Results[order.Pair]; !ok {
		c.Results[order.Pair] = &summary{Pair: order.Pair}
	}

	// register order volume
	c.Results[order.Pair].Volume += order.Price * order.Quantity

	// calculate profit only to sell orders
	if order.Side != model.SideTypeSell {
		return
	}

	profitValue, profit, err := c.calculateProfit(order)
	if err != nil {
		c.notifyError(err)
		return
	}

	order.Profit = profit
	if profitValue >= 0 {
		c.Results[order.Pair].Win = append(c.Results[order.Pair].Win, profitValue)
	} else {
		c.Results[order.Pair].Lose = append(c.Results[order.Pair].Lose, profitValue)
	}

	_, quote := exchange.SplitAssetQuote(order.Pair)
	c.notify(fmt.Sprintf("[PROFIT] %f %s (%f %%)\n`%s`", profitValue, quote, profit*100, c.Results[order.Pair].String()))
}

func (c *Controller) updateOrders() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// pending orders
	orders, err := c.storage.Orders(storage.WithStatusIn(
		model.OrderStatusTypeNew,
		model.OrderStatusTypePartiallyFilled,
		model.OrderStatusTypePendingCancel,
	))
	if err != nil {
		c.notifyError(err)
		c.mtx.Unlock()
		return
	}

	// For each pending order, check for updates
	var updatedOrders []model.Order
	for _, order := range orders {
		excOrder, err := c.exchange.Order(order.Pair, order.ExchangeID)
		if err != nil {
			log.WithField("id", order.ExchangeID).Error("orderController/get: ", err)
			continue
		}

		// no status change
		if excOrder.Status == order.Status {
			continue
		}

		excOrder.ID = order.ID
		err = c.storage.UpdateOrder(&excOrder)
		if err != nil {
			c.notifyError(err)
			continue
		}

		log.Infof("[ORDER %s] %s", excOrder.Status, excOrder)
		updatedOrders = append(updatedOrders, excOrder)
	}

	for _, processOrder := range updatedOrders {
		c.processTrade(&processOrder)
		c.monitor.Publish(processOrder)
	}
}

func (c *Controller) Status() Status {
	return c.status
}

func (c *Controller) Start() {
	if c.status != StatusRunning {
		c.status = StatusRunning
		go func() {
			ticker := time.NewTicker(c.tickerInterval)
			for {
				select {
				case <-ticker.C:
					c.updateOrders()
				case <-c.finish:
					ticker.Stop()
					return
				}
			}
		}()
		log.Info("Bot started.")
	}
}

func (c *Controller) Stop() {
	if c.status == StatusRunning {
		c.status = StatusStopped
		c.updateOrders()
		c.finish <- true
		log.Info("Bot stopped.")
	}
}

func (c *Controller) Account() (model.Account, error) {
	return c.exchange.Account()
}

func (c *Controller) Position(pair string) (asset, quote float64, err error) {
	return c.exchange.Position(pair)
}

func (c *Controller) LastQuote(pair string) (float64, error) {
	return c.exchange.GetLastQuote(c.ctx, pair)
}

func (c *Controller) PositionValue(pair string) (float64, error) {
	asset, _, err := c.exchange.Position(pair)
	if err != nil {
		return 0, err
	}
	return asset * c.lastPrice[pair], nil
}

func (c *Controller) Order(pair string, id int64) (model.Order, error) {
	return c.exchange.Order(pair, id)
}

func (c *Controller) CreateOrderOCO(side model.SideType, pair string, size, price, stop,
	stopLimit float64) ([]model.Order, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	log.Infof("[ORDER] Creating OCO order for %s", pair)
	orders, err := c.exchange.CreateOrderOCO(side, pair, size, price, stop, stopLimit)
	if err != nil {
		c.notifyError(err)
		return nil, err
	}

	for i := range orders {
		err := c.storage.CreateOrder(&orders[i])
		if err != nil {
			c.notifyError(err)
			return nil, err
		}
		go c.monitor.Publish(orders[i])
	}

	return orders, nil
}

func (c *Controller) CreateOrderLimit(side model.SideType, pair string, size, limit float64) (model.Order, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	log.Infof("[ORDER] Creating LIMIT %s order for %s", side, pair)
	order, err := c.exchange.CreateOrderLimit(side, pair, size, limit)
	if err != nil {
		c.notifyError(err)
		return model.Order{}, err
	}

	err = c.storage.CreateOrder(&order)
	if err != nil {
		c.notifyError(err)
		return model.Order{}, err
	}
	go c.monitor.Publish(order)
	log.Infof("[ORDER CREATED] %s", order)
	return order, nil
}

func (c *Controller) CreateOrderMarketQuote(side model.SideType, pair string, amount float64) (model.Order, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	log.Infof("[ORDER] Creating MARKET %s order for %s", side, pair)
	order, err := c.exchange.CreateOrderMarketQuote(side, pair, amount)
	if err != nil {
		c.notifyError(err)
		return model.Order{}, err
	}

	err = c.storage.CreateOrder(&order)
	if err != nil {
		c.notifyError(err)
		return model.Order{}, err
	}

	// calculate profit
	c.processTrade(&order)
	go c.monitor.Publish(order)
	log.Infof("[ORDER CREATED] %s", order)
	return order, err
}

func (c *Controller) CreateOrderMarket(side model.SideType, pair string, size float64) (model.Order, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	log.Infof("[ORDER] Creating MARKET %s order for %s", side, pair)
	order, err := c.exchange.CreateOrderMarket(side, pair, size)
	if err != nil {
		c.notifyError(err)
		return model.Order{}, err
	}

	err = c.storage.CreateOrder(&order)
	if err != nil {
		c.notifyError(err)
		return model.Order{}, err
	}

	// calculate profit
	c.processTrade(&order)
	go c.monitor.Publish(order)
	log.Infof("[ORDER CREATED] %s", order)
	return order, err
}

func (c *Controller) CreateOrderStop(pair string, size float64, limit float64) (model.Order, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	log.Infof("[ORDER] Creating STOP order for %s", pair)
	order, err := c.exchange.CreateOrderStop(pair, size, limit)
	if err != nil {
		c.notifyError(err)
		return model.Order{}, err
	}

	err = c.storage.CreateOrder(&order)
	if err != nil {
		c.notifyError(err)
		return model.Order{}, err
	}
	go c.monitor.Publish(order)
	log.Infof("[ORDER CREATED] %s", order)
	return order, nil
}

func (c *Controller) Cancel(order model.Order) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	log.Infof("[ORDER] Cancelling order for %s", order.Pair)
	err := c.exchange.Cancel(order)
	if err != nil {
		return err
	}

	order.Status = model.OrderStatusTypePendingCancel
	err = c.storage.UpdateOrder(&order)
	if err != nil {
		c.notifyError(err)
		return err
	}
	log.Infof("[ORDER CANCELED] %s", order)
	return nil
}
