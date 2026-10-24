package cryptodatadownload

// Ptr returns a pointer to v. Optional Params fields are pointers so that
// "not set" is explicit and the server applies its documented default;
// Ptr makes literals ergonomic:
//
//	rows, err := c.BinanceOHLC.Spot(ctx, cryptodatadownload.SpotParams{
//	    Symbol:   cryptodatadownload.Ptr("BTCUSDT"),
//	    Interval: cryptodatadownload.Ptr("1h"),
//	})
func Ptr[T any](v T) *T { return &v }
