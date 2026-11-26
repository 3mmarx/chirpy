# Terminal showcase for the README.
# Requires a running server on :8080 backed by a throwaway database.
# For a fresh database:
#   createdb chirpy_demo
#   goose -dir sql/schema postgres "postgres://localhost:5432/chirpy_demo?sslmode=disable" up
#   DB_URL="postgres://localhost:5432/chirpy_demo?sslmode=disable" \
#     JWT_SECRET=demo POLKA_KEY=demo PLATFORM=dev go run .
# Reset it before each render with:
#   curl -s -X POST localhost:8080/admin/reset
# Render from the repository root with:
#   vhs docs/demo.tape

Output docs/demo.gif

Set Shell "bash"
Set FontFamily "GoogleSansCode Nerd Font Mono"
Set FontSize 20
Set LineHeight 1.3
Set LetterSpacing 0.4
Set Width 1180
Set Height 620
Set Padding 16
Set Theme "Catppuccin Mocha"
Set WindowBar "Colorful"
Set WindowBarSize 40
Set BorderRadius 12
Set TypingSpeed 15ms
Set Framerate 30
Set PlaybackSpeed 1.0

Hide
Type `export PS1='❯ '`
Enter
Type "clear"
Enter
Show
Sleep 500ms

Type "API=localhost:8080"
Enter
Sleep 300ms

Type "curl -s $API/api/healthz; echo"
Enter
Sleep 800ms

Type `curl -s -X POST $API/api/users -H 'Content-Type: application/json' -d '{"email":"demo@chirpy.dev","password":"password"}' | jq '{email, is_chirpy_red}'`
Enter
Sleep 1500ms

Type `curl -s -X POST $API/api/login -H 'Content-Type: application/json' -d '{"email":"demo@chirpy.dev","password":"password"}' | jq '{email, is_chirpy_red}'`
Enter
Sleep 1500ms

Hide
Type `TOKEN=$(curl -s -X POST $API/api/login -H 'Content-Type: application/json' -d '{"email":"demo@chirpy.dev","password":"password"}' | jq -r .token)`
Enter
Type "clear"
Enter
Show
Sleep 100ms

Type `AUTH="Authorization: Bearer $TOKEN"`
Enter
Sleep 300ms

Type `curl -s -X POST $API/api/chirps -H "$AUTH" -H 'Content-Type: application/json' -d '{"body":"What a kerfuffle"}' | jq '{body}'`
Enter
Sleep 1500ms

Type `curl -s "$API/api/chirps?sort=desc" | jq '.[] | {body, created_at}'`
Enter
Sleep 2500ms
