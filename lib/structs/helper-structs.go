package structs

// FavoritesComparison holds the results of comparing local favorites with API
// favorites. It includes all favorites, exact matches, non-matches, conflicts,
// and local-only favorites. It's returned by the CombineFavorites function.
type FavoritesComparison struct {
	All       []Favorite // Combined local + API, deduplicated and deconflicted
	Exact     []Favorite // Exact matches (local + API)
	APIOnly   []Favorite // API-only favorites
	Conflicts []Favorite // Name conflicts (same name, different settings)
	LocalOnly []Favorite // Local-only favorites (not matched in API)
}
