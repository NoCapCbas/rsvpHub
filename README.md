# RSVP Hub for Baby Shower
This branch serves as actual working branch of service used to rsvp for Diaz Baby Shower

## Local Deployment

Install tailwind dependencies & Start tailwind engine
```bash
npm install
npx tailwindcss -i ./backend/src/input.css -o ./backend/src/output.css --watch
```

Navigate to backend directory & Start backend
```bash
go run main.go
```

Tailwind Config Content must match the following:
```javascript
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./backend/src/**/*.{html,js}"],
  theme: {
    extend: {},
  },
  plugins: [],
}

```

## Production Deployment
```bash
docker compose up --build
```
