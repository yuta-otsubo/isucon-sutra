package scenario

import (
	"context"
	"log/slog"

	"github.com/isucon/isucandar"
	"github.com/samber/lo"
	"github.com/yuta-otsubo/isucon-sutra/bench/benchmarker/world"
)

// Validation はシナリオの結果検証処理を行う
func (s *Scenario) Validation(ctx context.Context, step *isucandar.BenchmarkStep) error {
	actual := s.world.PaymentDB.TotalPayment() + s.TotalDiscount()
	expected := s.TotalSales()
	if actual != expected {
		s.contestantLogger.Error("決済サーバーで決済された額とユーザーが支払うべき額が一致していません", slog.Int64("diff(actual-expected)", actual-expected))
	}

	for _, region := range s.world.Regions {
		s.contestantLogger.Info("最終Region情報",
			slog.String("region", region.Name),
			slog.Int("users", region.UsersDB.Len()),
			slog.Int("active_users", region.ActiveUserNum()),
		)
	}
	for id, owner := range s.world.OwnerDB.Iter() {
		s.contestantLogger.Info("最終Owner情報",
			slog.Int("id", int(id)),
			slog.Int64("total_sales", owner.TotalSales.Load()),
			slog.Int("chairs", owner.ChairDB.Len()),
			slog.Int("chairs_outside_region", lo.CountBy(owner.ChairDB.ToSlice(), func(c *world.Chair) bool { return !c.Location.Current().Within(owner.Region) })),
			slog.Int("total_chair_travel_distance", lo.SumBy(owner.ChairDB.ToSlice(), func(c *world.Chair) int { return c.Location.TotalTravelDistance() })),
		)
	}
	s.contestantLogger.Info("種別エラー発生数", slog.Any("errors", s.world.ErrorCounter.Count()))
	return sendResult(s, true, true)
}
