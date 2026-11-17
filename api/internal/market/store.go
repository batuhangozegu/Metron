package market

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Postgres struct{ Pool *pgxpool.Pool }

func (p Postgres) Symbols(ctx context.Context) ([]Symbol, error) {
	rows, err := p.Pool.Query(ctx, `SELECT symbol, base_asset, quote_asset, coin_class,
		atr_1d::text, atr_1d_percent::text FROM market.symbols
		ORDER BY avg_daily_quote_volume DESC NULLS LAST, symbol`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Symbol
	for rows.Next() {
		var row Symbol
		if err := rows.Scan(&row.Symbol, &row.BaseAsset, &row.QuoteAsset, &row.CoinClass, &row.ATR, &row.ATRPercent); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

type Redis struct{ Client *redis.Client }

func (r Redis) Tickers(ctx context.Context, symbols []string) (map[string]map[string]string, error) {
	commands := make(map[string]*redis.MapStringStringCmd, len(symbols))
	_, err := r.Client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, symbol := range symbols {
			commands[symbol] = pipe.HGetAll(ctx, "market:ticker:"+symbol)
		}
		return nil
	})
	result := make(map[string]map[string]string, len(commands))
	for symbol, command := range commands {
		if values, cmdErr := command.Result(); cmdErr == nil {
			result[symbol] = values
		}
	}
	return result, err
}
