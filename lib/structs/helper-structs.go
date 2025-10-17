package structs

// FavoritesComparison holds the results of comparing local favorites with API
// favorites. It includes all favorites, exact matches, non-matches, conflicts,
// and local-only favorites. It's returned by the CombineFavorites function.
type FavoritesComparison struct {
	All               []Favorite // Combined local + API, deduplicated and deconflicted
	ConflictsLocal    []Favorite // Name conflicts (same name, different settings)
	ConflictsUpstream []Favorite // Name conflicts (same name, different settings)
	LocalOnly         []Favorite // Local-only favorites (not matched in API)
	UnaliasedLocal    []Favorite // Local favorites that update unnamed API favorites
	UnaliasedUpstream []Favorite // Local favorites that update unnamed API favorites
}
