package entity

type LimiterRepository interface {
	Create(limiter *Limiter) error
	Update(limiter *Limiter) error
	Find(id string) (*Limiter, error)
}
