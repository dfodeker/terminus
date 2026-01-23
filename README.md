# StoreOS

An open-source e-commerce infrastructure for developers who want to build, experiment, and ship — without the paywall.

## Motivation

I've spent a lot of time building with Shopify. It's powerful, but every time I wanted to spin up a headless storefront or experiment with a new idea, I hit the same wall: *"Choose a plan to continue."* (not the exact words, but you get the idea).

I just wanted to play around. Test an idea. Build something.

StoreOS is the shopping infrastructure I wished existed — one that stays conceptually close to Shopify (products, variants, collections, themes) but gets out of your way. It's open-source, self-hostable, and built for developers who want a commerce backend they can actually tinker with.

Think of it as a cross between Shopify's mental model and Medusa's developer-first philosophy, with a few ideas of its own — like built-in feature flags so merchants can incrementally roll out changes to customers instead of the all-or-nothing deployments we've all grown to dread.

## Features

- **Multi-tenant architecture** — Organizations, stores, and role-based access control
- **Products & variants** — Flexible product modeling
- **Feature flags** — Roll out storefront changes incrementally (coming soon)
- **Themes** — Customizable storefronts with a lightweight templating system (coming soon)
- **Headless-first** — Build any frontend you want

### Roadmap

- [ ] Collections
- [ ] Discounts
- [ ] Checkout with Stripe
- [ ] Theme engine (Liquid-inspired)
- [ ] Admin dashboard

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+

### Installation

```bash
git clone https://github.com/dfodeker/storeOS.git
cd storeOS
```

### Configuration

```bash
cp .env.example .env
```

Set your environment variables:

```env
DATABASE_URL=postgres://user:password@localhost:5432/storeos?sslmode=disable
PORT=8080
JWT_SECRET=your-secret-key
```

### Run

```bash
# Run migrations
make migrate

# Start the server
make run
```

The API will be available at `http://localhost:8080`.

## Project Structure

```
/cmd
    /api                 # Application entrypoint
/internal
    /domain              # Core business entities and interfaces
    /application         # Use cases and services  
    /adapter
        /postgres        # Database implementations
        /http            # Handlers and middleware
/migrations              # SQL migrations
/config                  # Configuration loading
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go |
| Database | PostgreSQL |
| Migrations | goose |
| SQL | sqlc |
| Auth | JWT |

## Contributing

StoreOS is early and evolving. If you're into e-commerce infra or just want to poke around, contributions are welcome.

1. Fork the repo
2. Create your feature branch (`git checkout -b feature/cool-thing`)
3. Commit your changes (`git commit -m 'Add cool thing'`)
4. Push to the branch (`git push origin feature/cool-thing`)
5. Open a Pull Request

## License

MIT