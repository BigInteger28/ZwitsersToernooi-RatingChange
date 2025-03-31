package main

import (
    "bufio"
    "fmt"
    "html/template"
    "os"
    "sort"
    "strconv"
    "strings"
)

// Player struct om een speler te vertegenwoordigen
type Player struct {
    Name         string
    Level        int
    Rating       int
    Punten       int     // 2 voor winst, 1 voor gelijkspel, 0 voor verlies
    Matchscore   float64 // Cumulatieve scoreverschillen: eigen score - score tegenstander
    Opponents    []string
    RatOppTotal  float64 // Totale som van ratings van tegenstanders
    RoundsPlayed int     // Aantal gespeelde rondes
}

// Match struct voor een pairing
type Match struct {
    Player1 Player
    Player2 Player
    Result  string // bv "3-3"
}

// Result struct voor scores uit rondeX.txt
type Result struct {
    Player1 string
    Player2 string
    Score1  float64
    Score2  float64
}

// Spelers inlezen uit input.txt
func readPlayers(filename string) ([]Player, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var players []Player
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.Split(line, "   ") // Drie spaties
        if len(parts) != 3 {
            continue
        }
        level, _ := strconv.Atoi(parts[1])
        rating, _ := strconv.Atoi(parts[2])
        players = append(players, Player{
            Name:         parts[0],
            Level:        level,
            Rating:       rating,
            Punten:       0,
            Matchscore:   0.0,
            Opponents:    []string{},
            RatOppTotal:  0.0,
            RoundsPlayed: 0,
        })
    }
    return players, scanner.Err()
}

// Matches laden uit rondeX.txt
func loadMatches(filename string, players []Player) ([]Match, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var matches []Match
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.Split(line, "   ")
        if len(parts) != 3 {
            continue
        }
        p1Name := strings.Split(parts[0], " LVL")[0]
        p2Name := strings.Split(parts[2], " LVL")[0]
        if p2Name == "Bye" {
            p2Name = "Bye"
        }
        result := parts[1]

        var p1, p2 Player
        for _, player := range players {
            if player.Name == p1Name {
                p1 = player
            } else if player.Name == p2Name {
                p2 = player
            }
        }
        matches = append(matches, Match{Player1: p1, Player2: p2, Result: result})
    }
    return matches, scanner.Err()
}

func savePlayerStatus(filename string, players []Player) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    for _, player := range players {
        line := fmt.Sprintf("%s,%d,%d,%d,%.2f,%.2f,%d,%s\n",
            player.Name, player.Level, player.Rating, player.Punten,
            player.Matchscore, player.RatOppTotal, player.RoundsPlayed,
            strings.Join(player.Opponents, ";"))
        file.WriteString(line)
    }
    return nil
}

func loadPlayerStatus(filename string, players []Player) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.Split(line, ",")
        if len(parts) != 8 {
            continue
        }
        name := parts[0]
        level, _ := strconv.Atoi(parts[1])
        rating, _ := strconv.Atoi(parts[2])
        punten, _ := strconv.Atoi(parts[3])
        matchscore, _ := strconv.ParseFloat(parts[4], 64)
        ratOppTotal, _ := strconv.ParseFloat(parts[5], 64)
        roundsPlayed, _ := strconv.Atoi(parts[6])
        opponents := strings.Split(parts[7], ";")

        for i := range players {
            if players[i].Name == name {
                players[i].Level = level
                players[i].Rating = rating
                players[i].Punten = punten
                players[i].Matchscore = matchscore
                players[i].RatOppTotal = ratOppTotal
                players[i].RoundsPlayed = roundsPlayed
                players[i].Opponents = opponents
                break
            }
        }
    }
    return scanner.Err()
}

// Spelers sorteren op Punten, dan Matchscore, dan RatOpp, dan Rating (allemaal aflopend)
func sortPlayers(players []Player) {
    sort.Slice(players, func(i, j int) bool {
        if players[i].Punten != players[j].Punten {
            return players[i].Punten > players[j].Punten
        }
        if players[i].Matchscore != players[j].Matchscore {
            return players[i].Matchscore > players[j].Matchscore
        }
        var ratOppI, ratOppJ float64
        if players[i].RoundsPlayed > 0 {
            ratOppI = players[i].RatOppTotal / float64(players[i].RoundsPlayed)
        }
        if players[j].RoundsPlayed > 0 {
            ratOppJ = players[j].RatOppTotal / float64(players[j].RoundsPlayed)
        }
        if ratOppI != ratOppJ {
            return ratOppI > ratOppJ
        }
        return players[i].Rating > players[j].Rating
    })
}

// Check of twee spelers al tegen elkaar hebben gespeeld
func hasPlayed(p1, p2 Player) bool {
    for _, opp := range p1.Opponents {
        if opp == p2.Name {
            return true
        }
    }
    return false
}

// Pairings maken voor een ronde met prioriteit voor nieuwe tegenstanders met dezelfde score
func pairPlayers(players []Player) []Match {
    sortPlayers(players)
    var matches []Match
    used := make(map[string]bool)

    // Groepeer spelers per score
    scoreGroups := make(map[int][]Player)
    for _, p := range players {
        if !used[p.Name] {
            scoreGroups[p.Punten] = append(scoreGroups[p.Punten], p)
        }
    }

    // Pair spelers binnen scoregroepen
    for score := range scoreGroups {
        group := scoreGroups[score]
        i := 0
        for i < len(group) {
            p1 := group[i]
            if used[p1.Name] {
                i++
                continue
            }
            paired := false
            for j := i + 1; j < len(group); j++ {
                p2 := group[j]
                if !used[p2.Name] && !hasPlayed(p1, p2) {
                    matches = append(matches, Match{Player1: p1, Player2: p2, Result: "0-0"})
                    used[p1.Name] = true
                    used[p2.Name] = true
                    paired = true
                    break
                }
            }
            if paired {
                // Verwijder gepairde spelers uit de groep
                newGroup := []Player{}
                for _, p := range group {
                    if !used[p.Name] {
                        newGroup = append(newGroup, p)
                    }
                }
                group = newGroup
            } else {
                i++
            }
        }
        scoreGroups[score] = group
    }

    // Verzamel overgebleven spelers
    var leftovers []Player
    for _, group := range scoreGroups {
        for _, p := range group {
            if !used[p.Name] {
                leftovers = append(leftovers, p)
            }
        }
    }

    // Fase 1: Pair leftovers zonder herhalingen
    i := 0
    for i < len(leftovers) {
        p1 := leftovers[i]
        if used[p1.Name] {
            i++
            continue
        }
        paired := false
        for j := i + 1; j < len(leftovers); j++ {
            p2 := leftovers[j]
            if !used[p2.Name] && !hasPlayed(p1, p2) {
                matches = append(matches, Match{Player1: p1, Player2: p2, Result: "0-0"})
                used[p1.Name] = true
                used[p2.Name] = true
                paired = true
                break
            }
        }
        if paired {
            i = 0 // Reset om opnieuw te beginnen
        } else {
            i++
        }
    }

    // Fase 2: Pair overgebleven spelers, herhalingen toegestaan
    remaining := []Player{}
    for _, p := range leftovers {
        if !used[p.Name] {
            remaining = append(remaining, p)
        }
    }
    for i := 0; i < len(remaining); i += 2 {
        if i+1 < len(remaining) {
            p1 := remaining[i]
            p2 := remaining[i+1]
            matches = append(matches, Match{Player1: p1, Player2: p2, Result: "0-0"})
            used[p1.Name] = true
            used[p2.Name] = true
        }
    }

    return matches
}

// RondeX.txt genereren
func generateRoundFile(round int, matches []Match) error {
    filename := fmt.Sprintf("ronde%d.txt", round)
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    for _, match := range matches {
        if match.Player2.Name == "Bye" {
            line := fmt.Sprintf("%s LVL %d (%d rating)   1-0   Bye\n",
                match.Player1.Name, match.Player1.Level, match.Player1.Rating)
            file.WriteString(line)
        } else {
            line := fmt.Sprintf("%s LVL %d (%d rating)   0-0   %s LVL %d (%d rating)\n",
                match.Player1.Name, match.Player1.Level, match.Player1.Rating,
                match.Player2.Name, match.Player2.Level, match.Player2.Rating)
            file.WriteString(line)
        }
    }
    return nil
}

// Scores inlezen uit rondeX.txt
func readRoundResults(filename string) ([]Result, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var results []Result
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.Split(line, "   ")
        if len(parts) != 3 {
            continue
        }
        p1Name := strings.Split(parts[0], " LVL")[0]
        p2Name := strings.Split(parts[2], " LVL")[0]
        if p2Name == "Bye" {
            p2Name = "Bye"
        }
        scores := strings.Split(parts[1], "-")
        if len(scores) != 2 {
            continue
        }
        score1, _ := strconv.ParseFloat(scores[0], 64)
        score2, _ := strconv.ParseFloat(scores[1], 64)
        results = append(results, Result{
            Player1: p1Name,
            Player2: p2Name,
            Score1:  score1,
            Score2:  score2,
        })
    }
    return results, scanner.Err()
}

// Spelers updaten met Punten, Matchscore en RatOpp
func updatePlayers(players []Player, results []Result) {
    for _, result := range results {
        for i := range players {
            if players[i].Name == result.Player1 {
                // Update Punten
                if result.Score1 > result.Score2 {
                    players[i].Punten += 2
                } else if result.Score1 == result.Score2 {
                    players[i].Punten += 1
                }
                // Update Matchscore: start met vorige Matchscore en tel (eigen score - score tegenstander) op
                players[i].Matchscore += (result.Score1 - result.Score2)
                if result.Player2 != "Bye" {
                    players[i].Opponents = append(players[i].Opponents, result.Player2)
                    // Voeg rating van tegenstander toe aan RatOppTotal
                    for _, opp := range players {
                        if opp.Name == result.Player2 {
                            players[i].RatOppTotal += float64(opp.Rating)
                            break
                        }
                    }
                    players[i].RoundsPlayed++
                }
            } else if players[i].Name == result.Player2 && result.Player2 != "Bye" {
                // Update Punten
                if result.Score2 > result.Score1 {
                    players[i].Punten += 2
                } else if result.Score2 == result.Score1 {
                    players[i].Punten += 1
                }
                // Update Matchscore: start met vorige Matchscore en tel (eigen score - score tegenstander) op
                players[i].Matchscore += (result.Score2 - result.Score1)
                players[i].Opponents = append(players[i].Opponents, result.Player1)
                // Voeg rating van tegenstander toe aan RatOppTotal
                for _, opp := range players {
                    if opp.Name == result.Player1 {
                        players[i].RatOppTotal += float64(opp.Rating)
                        break
                    }
                }
                players[i].RoundsPlayed++
            }
        }
    }
}

// Update match results
func updateMatchResults(matches []Match, results []Result) {
    for i, match := range matches {
        for _, result := range results {
            if match.Player1.Name == result.Player1 && match.Player2.Name == result.Player2 {
                matches[i].Result = fmt.Sprintf("%v-%v", result.Score1, result.Score2)
                break
            }
        }
    }
}

// HTML genereren met CSS voor centrering, randen en padding
func generateHTML(round int, players []Player, matches []Match) error {
    const tmpl = `
    <html>
    <head>
    <title>Ronde {{.Round}}</title>
    <style>
    body {
        text-align: center;
    }
    table {
        border-collapse: collapse;
        margin: auto;
    }
    table, th, td {
        border: 1px solid lightgray;
        text-align: center;
        padding: 5px;
    }
    </style>
    </head>
    <body>
    <h1>Ronde {{.Round}}</h1>
    <h2>Standings</h2>
    <table>
        <tr>
            <th>Nr.</th>
            <th>Naam</th>
            <th>Level</th>
            <th>Rating</th>
            <th>Punten</th>
            <th>Matchscore</th>
            <th>RatOpp</th>
        </tr>
        {{range $index, $player := .Players}}
        <tr>
            <td>{{add $index 1}}</td>
            <td>{{$player.Name}}</td>
            <td>{{$player.Level}}</td>
            <td>{{$player.Rating}}</td>
            <td>{{$player.Punten}}</td>
            <td>{{$player.Matchscore}}</td>
            <td>{{if $player.RoundsPlayed}}{{printf "%.2f" (div $player.RatOppTotal $player.RoundsPlayed)}}{{else}}0{{end}}</td>
        </tr>
        {{end}}
    </table>
    <h2>Pairings</h2>
    <table>
        <tr>
            <th>Nr.</th>
            <th>Naam</th>
            <th>Level</th>
            <th>Rating</th>
            <th>Score</th>
            <th>Naam</th>
            <th>Level</th>
            <th>Rating</th>
        </tr>
        {{range $index, $match := .Matches}}
        <tr>
            <td>{{add $index 1}}</td>
            <td>{{$match.Player1.Name}}</td>
            <td>{{$match.Player1.Level}}</td>
            <td>{{$match.Player1.Rating}}</td>
            <td>{{$match.Result}}</td>
            <td>{{$match.Player2.Name}}</td>
            {{if eq $match.Player2.Name "Bye"}}
            <td>-</td>
            <td>-</td>
            {{else}}
            <td>{{$match.Player2.Level}}</td>
            <td>{{$match.Player2.Rating}}</td>
            {{end}}
        </tr>
        {{end}}
    </table>
    </body>
    </html>`

    t := template.Must(template.New("round").Funcs(template.FuncMap{
        "add": func(a int, b int) int { return a + b },
        "div": func(a float64, b int) float64 {
            if b == 0 {
                return 0 // Voorkomt deling door nul
            }
            return a / float64(b)
        },
    }).Parse(tmpl))

    filename := fmt.Sprintf("ronde%d.html", round)
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    sortPlayers(players)
    data := struct {
        Round   int
        Players []Player
        Matches []Match
    }{Round: round, Players: players, Matches: matches}
    return t.Execute(file, data)
}

// RATING BEREKENING SPELERS
func getBonus(theRange int, maxRatingAdd int, ratingOpponent int, ownRating int, result string) int {
    var bonus int
    perRating := (2 * theRange) / maxRatingAdd
    low := ownRating - theRange
    if ratingOpponent <= ownRating-theRange {
        bonus = 0
    } else if ratingOpponent >= ownRating+theRange {
        bonus = maxRatingAdd
    } else {
        bonus = (ratingOpponent - low) / perRating
    }
    if result == "w" {
        return bonus
    } else if result == "d" {
        if ratingOpponent >= ownRating {
            return bonus / 2
        } else {
            return 0 - ((maxRatingAdd - bonus) / 2)
        }
    } else {
        return 0 - (maxRatingAdd - bonus)
    }
}

func getMatchOutcome(playerName string, result Result) string {
    if playerName == result.Player1 {
        if result.Score1 > result.Score2 {
            return "w"
        } else if result.Score1 == result.Score2 {
            return "d"
        } else {
            return "l"
        }
    } else if playerName == result.Player2 {
        if result.Score2 > result.Score1 {
            return "w"
        } else if result.Score2 == result.Score1 {
            return "d"
        } else {
            return "l"
        }
    }
    return ""
}

func outcomeToString(outcome string) string {
    switch outcome {
    case "w":
        return "WIN"
    case "d":
        return "DRAW"
    case "l":
        return "LOSE"
    default:
        return ""
    }
}

func generateRatingHTML(players []Player, allResults [][]Result, initialRatings map[string]int) error {
    var theRange int = 450
    var maxRatingAdd int = 30

    var htmlContent strings.Builder
    htmlContent.WriteString(`
    <html>
    <head>
    <style>
    body {
        text-align: center;
    }
    table {
        border-collapse: collapse;
        margin: auto;
    }
    th, td {
        border: 1px solid lightgray;
        padding: 10px;
        text-align: center;
    }
    </style>
    </head>
    <body>
    `)

    for _, player := range players {
        totalAdd := 0
        htmlContent.WriteString(fmt.Sprintf("<p>%s - Level %d</p>", player.Name, player.Level))
        htmlContent.WriteString(fmt.Sprintf("<p>EIGEN RATING START: %d</p>", initialRatings[player.Name]))
        htmlContent.WriteString("<table><tr><th>Naam</th><th>Level</th><th>Rating</th><th>Resultaat</th><th>Rating erbij</th></tr>")

        for _, results := range allResults {
            for _, result := range results {
                if result.Player1 == player.Name || result.Player2 == player.Name {
                    var opponentName string
                    var opponentRating, opponentLevel int
                    var outcome string
                    if result.Player1 == player.Name {
                        opponentName = result.Player2
                        outcome = getMatchOutcome(player.Name, result)
                        for _, p := range players {
                            if p.Name == opponentName {
                                opponentRating = initialRatings[p.Name]
                                opponentLevel = p.Level
                                break
                            }
                        }
                    } else {
                        opponentName = result.Player1
                        outcome = getMatchOutcome(player.Name, result)
                        for _, p := range players {
                            if p.Name == opponentName {
                                opponentRating = initialRatings[p.Name]
                                opponentLevel = p.Level
                                break
                            }
                        }
                    }
                    if opponentName != "Bye" {
                        // Gebruik altijd de initiële rating, niet een bijgewerkte versie
                        bonus := getBonus(theRange, maxRatingAdd, opponentRating, initialRatings[player.Name], outcome)
                        totalAdd += bonus
                        htmlContent.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%d</td><td>%d</td><td>%s</td><td>%+d</td></tr>",
                            opponentName, opponentLevel, opponentRating, outcomeToString(outcome), bonus))
                    }
                }
            }
        }
        htmlContent.WriteString("</table>")
        htmlContent.WriteString(fmt.Sprintf("<p>RATING ERBIJ: %+d</p>", totalAdd))
        htmlContent.WriteString(fmt.Sprintf("<p>NIEUWE RATING: %d</p>", initialRatings[player.Name]+totalAdd))
        htmlContent.WriteString("<hr>")
    }

    htmlContent.WriteString("</body></html>")

    file, err := os.Create("rating_update.html")
    if err != nil {
        return err
    }
    defer file.Close()
    file.WriteString(htmlContent.String())
    return nil
}

// Hoofdprogramma met menu
func main() {
    players, err := readPlayers("input.txt")
    if err != nil {
        fmt.Println("Fout bij inlezen spelers:", err)
        return
    }

    currentRound := 0
    var lastMatches []Match

    for {
        fmt.Println("\nMenu:")
        fmt.Println("0. Verander huidige rondenr")
        fmt.Println("1. Genereer nieuwe ronde")
        fmt.Println("2. Genereer finale ronde")
		fmt.Println("3. Verwerk scores van huidige ronde")
        fmt.Println("4. Genereer HTML")        
        fmt.Println("5. Verwerk nieuwe rating alle spelers")
        fmt.Println("6. Exit")
        fmt.Print("Kies een optie: ")

        var choice string
        fmt.Scanln(&choice)

        switch choice {
        case "0":
            fmt.Print("Voer nieuwe rondenr in: ")
            var newRound string
            fmt.Scanln(&newRound)
            if roundNum, err := strconv.Atoi(newRound); err == nil {
                currentRound = roundNum
                fmt.Println("Huidige rondenr is nu:", currentRound)

                // Laad spelerstatus
                statusFile := fmt.Sprintf("ronde%d_status.txt", currentRound)
                if _, err := os.Stat(statusFile); err == nil {
                    if err := loadPlayerStatus(statusFile, players); err != nil {
                        fmt.Println("Fout bij laden spelerstatus:", err)
                    } else {
                        fmt.Println("Spelerstatus geladen voor ronde", currentRound)
                    }
                } else {
                    fmt.Println("Geen spelerstatus gevonden voor ronde", currentRound, "- start met schone lei")
                }

                // Laad matches
                filename := fmt.Sprintf("ronde%d.txt", currentRound)
                if _, err := os.Stat(filename); err == nil {
                    lastMatches, err = loadMatches(filename, players)
                    if err != nil {
                        fmt.Println("Fout bij laden matches:", err)
                    } else {
                        fmt.Println("Matches geladen voor ronde", currentRound)
                    }
                } else {
                    fmt.Println("Geen matches gevonden voor ronde", currentRound)
                    lastMatches = nil // Reset matches als er geen bestand is
                }
            } else {
                fmt.Println("Ongeldig rondenr")
            }

        case "1":
            currentRound++
            lastMatches = pairPlayers(players)
            if err := generateRoundFile(currentRound, lastMatches); err != nil {
                fmt.Println("Fout bij genereren ronde:", err)
            } else {
                fmt.Printf("Ronde %d gegenereerd. Vul de scores in in ronde%d.txt\n", currentRound, currentRound)
            }

        case "2":
            sortPlayers(players)
            if len(players) < 2 {
                fmt.Println("Niet genoeg spelers voor finale")
                continue
            }
            currentRound++
            finalMatch := Match{Player1: players[0], Player2: players[1], Result: "0-0"}
            lastMatches = []Match{finalMatch}
            if err := generateRoundFile(currentRound, lastMatches); err != nil {
                fmt.Println("Fout bij genereren finale ronde:", err)
            } else {
                fmt.Println("Finale ronde gegenereerd.")
            }
			
        case "3":
            filename := fmt.Sprintf("ronde%d.txt", currentRound)
            results, err := readRoundResults(filename)
            if err != nil {
                fmt.Println("Fout bij inlezen scores:", err)
            } else {
                updatePlayers(players, results) // Werk spelerstatistieken bij
                updateMatchResults(lastMatches, results) // Werk matches bij
                statusFile := fmt.Sprintf("ronde%d_status.txt", currentRound)
                if err := savePlayerStatus(statusFile, players); err != nil {
                    fmt.Println("Fout bij opslaan spelerstatus:", err)
                } else {
                    fmt.Println("Spelerstatus opgeslagen voor ronde", currentRound)
                }
                fmt.Println("Scores verwerkt voor ronde", currentRound)
            }

        case "4":
            if len(lastMatches) == 0 {
                fmt.Println("Geen matches beschikbaar om HTML te genereren. Genereer eerst een ronde of laad de matches.")
            } else if err := generateHTML(currentRound, players, lastMatches); err != nil {
                fmt.Println("Fout bij genereren HTML:", err)
            } else {
                fmt.Println("HTML gegenereerd voor ronde", currentRound)
            }

        case "5":
            // Verzamel alle resultaten van alle rondes
            var allResults [][]Result
            for r := 1; r <= currentRound; r++ {
                filename := fmt.Sprintf("ronde%d.txt", r)
                results, err := readRoundResults(filename)
                if err != nil {
                    fmt.Println("Fout bij inlezen results voor ronde", r, ":", err)
                    continue
                }
                allResults = append(allResults, results)
            }
            // Verzamel initiële ratings
            initialRatings := make(map[string]int)
            for _, p := range players {
                initialRatings[p.Name] = p.Rating
            }
            if err := generateRatingHTML(players, allResults, initialRatings); err != nil {
                fmt.Println("Fout bij genereren rating HTML:", err)
            } else {
                fmt.Println("Rating update HTML gegenereerd in 'rating_update.html'")
            }

        case "6":
            fmt.Println("Exit")
            os.Exit(0)

        default:
            fmt.Println("Ongeldige keuze")
        }
    }
}
