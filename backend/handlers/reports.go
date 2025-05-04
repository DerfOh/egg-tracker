package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/marcboeker/go-duckdb"
)

// ReportsHandler returns analytics from DuckDB for the reports page.
func ReportsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("[ReportsHandler] Opening DuckDB connection...")
		duckdb, err := sql.Open("duckdb", "/app/data/eggtracker.duckdb")
		if err != nil {
			log.Printf("[ReportsHandler] Failed to open DuckDB: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open DuckDB"})
			return
		}
		defer duckdb.Close()

		// 1. Eggs over time by species (from inventory_actions, action = 'collected')
		log.Println("[ReportsHandler] Querying eggs over time by species from inventory_actions...")
		eggsRows, err := duckdb.Query(`
			SELECT date, species, SUM(quantity) as count
			FROM inventory_actions
			WHERE action = 'collected'
			GROUP BY date, species
			ORDER BY date ASC
		`)
		if err != nil {
			log.Printf("[ReportsHandler] Eggs query failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "eggs query failed"})
			return
		}
		eggsMap := map[string]map[string]int{}
		rowCount := 0
		for eggsRows.Next() {
			var date, species string
			var count int
			eggsRows.Scan(&date, &species, &count)
			if _, ok := eggsMap[date]; !ok {
				eggsMap[date] = map[string]int{}
			}
			eggsMap[date][species] = count
			rowCount++
		}
		log.Printf("[ReportsHandler] Eggs query returned %d rows", rowCount)
		eggsRows.Close()
		var eggsOverTime []map[string]interface{}
		for date, speciesCounts := range eggsMap {
			row := map[string]interface{}{"date": date}
			for sp, cnt := range speciesCounts {
				row[sp] = cnt
			}
			eggsOverTime = append(eggsOverTime, row)
		}

		// 2. Inventory trends by action
		log.Println("[ReportsHandler] Querying inventory trends by action...")
		invRows, err := duckdb.Query(`
			SELECT date, action, SUM(quantity) as qty
			FROM inventory_actions
			GROUP BY date, action
			ORDER BY date ASC
		`)
		if err != nil {
			log.Printf("[ReportsHandler] Inventory query failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "inventory query failed"})
			return
		}
		invMap := map[string]map[string]int{}
		invRowCount := 0
		for invRows.Next() {
			var date, action string
			var qty int
			invRows.Scan(&date, &action, &qty)
			if _, ok := invMap[date]; !ok {
				invMap[date] = map[string]int{}
			}
			invMap[date][action] = qty
			invRowCount++
		}
		log.Printf("[ReportsHandler] Inventory query returned %d rows", invRowCount)
		invRows.Close()
		var inventoryTrends []map[string]interface{}
		for date, actions := range invMap {
			row := map[string]interface{}{"date": date}
			for act, qty := range actions {
				row[act] = qty
			}
			inventoryTrends = append(inventoryTrends, row)
		}

		// 3. Average eggs/day per coop (from inventory_actions, action = 'collected')
		log.Println("[ReportsHandler] Querying average eggs/day per coop from inventory_actions...")
		avgRows, err := duckdb.Query(`
			SELECT coop, AVG(cnt) as avg
			FROM (
				SELECT coop, date, SUM(quantity) as cnt
				FROM inventory_actions
				WHERE action = 'collected'
				GROUP BY coop, date
			)
			GROUP BY coop
		`)
		if err != nil {
			log.Printf("[ReportsHandler] Avg eggs/coop query failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "avg eggs/coop query failed"})
			return
		}
		var avgEggsPerCoop []map[string]interface{}
		avgRowCount := 0
		for avgRows.Next() {
			var coop string
			var avg float64
			avgRows.Scan(&coop, &avg)
			avgEggsPerCoop = append(avgEggsPerCoop, map[string]interface{}{
				"coop": coop,
				"avg":  avg,
			})
			avgRowCount++
		}
		log.Printf("[ReportsHandler] Avg eggs/coop query returned %d rows", avgRowCount)
		avgRows.Close()

		// 4. Eggs collected by week (all species)
		log.Println("[ReportsHandler] Querying eggs collected by week...")
		weeklyRows, err := duckdb.Query(`
			SELECT strftime(date, '%Y-%W') as week, SUM(quantity) as count
			FROM inventory_actions
			WHERE action = 'collected'
			GROUP BY week
			ORDER BY week ASC
		`)
		if err != nil {
			log.Printf("[ReportsHandler] Weekly eggs query failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "weekly eggs query failed"})
			return
		}
		var eggsByWeek []map[string]interface{}
		for weeklyRows.Next() {
			var week string
			var count int
			weeklyRows.Scan(&week, &count)
			eggsByWeek = append(eggsByWeek, map[string]interface{}{
				"week":  week,
				"count": count,
			})
		}
		weeklyRows.Close()

		// 5. Inventory actions by species
		log.Println("[ReportsHandler] Querying inventory actions by species...")
		speciesRows, err := duckdb.Query(`
			SELECT species, action, SUM(quantity) as qty
			FROM inventory_actions
			GROUP BY species, action
		`)
		if err != nil {
			log.Printf("[ReportsHandler] Inventory by species query failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "inventory by species query failed"})
			return
		}
		speciesActionMap := map[string]map[string]int{}
		for speciesRows.Next() {
			var species, action string
			var qty int
			speciesRows.Scan(&species, &action, &qty)
			if _, ok := speciesActionMap[species]; !ok {
				speciesActionMap[species] = map[string]int{}
			}
			speciesActionMap[species][action] = qty
		}
		speciesRows.Close()
		var inventoryBySpecies []map[string]interface{}
		for species, actions := range speciesActionMap {
			row := map[string]interface{}{"species": species}
			for act, qty := range actions {
				row[act] = qty
			}
			inventoryBySpecies = append(inventoryBySpecies, row)
		}

		// Compute net totals for each species
		netTotals := []map[string]interface{}{}
		for species, actions := range speciesActionMap {
			net := 0
			net += actions["collected"]
			net -= actions["sold"]
			net -= actions["consumed"]
			net -= actions["gifted"]
			net -= actions["spoiled"]
			row := map[string]interface{}{"species": species, "net": net}
			netTotals = append(netTotals, row)
		}

		// 6. Top producing species (total eggs collected per species)
		log.Println("[ReportsHandler] Querying top producing species...")
		topRows, err := duckdb.Query(`
			SELECT species, SUM(quantity) as total
			FROM inventory_actions
			WHERE action = 'collected'
			GROUP BY species
			ORDER BY total DESC
		`)
		if err != nil {
			log.Printf("[ReportsHandler] Top species query failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "top species query failed"})
			return
		}
		var topSpecies []map[string]interface{}
		for topRows.Next() {
			var species string
			var total int
			topRows.Scan(&species, &total)
			topSpecies = append(topSpecies, map[string]interface{}{
				"species": species,
				"total":   total,
			})
		}
		topRows.Close()

		log.Println("[ReportsHandler] Returning JSON response.")
		c.JSON(http.StatusOK, gin.H{
			"eggsOverTime":       eggsOverTime,
			"inventoryTrends":    inventoryTrends,
			"avgEggsPerCoop":     avgEggsPerCoop,
			"eggsByWeek":         eggsByWeek,
			"inventoryBySpecies": inventoryBySpecies,
			"topSpecies":         topSpecies,
			"netTotals":          netTotals,
		})
	}
}
