# 🤝 Contributors Guide

Thanks for your interest in contributing to **bffgen**! 🎉
This project is being built in the open, and contributions of all kinds are welcome — from code to docs to research into developer experience (DX).

---

## 🚀 How to Contribute

### 1. Fork & Clone

```bash
git clone https://github.com/YOUR-USERNAME/bffgen.git
cd bffgen
```

### 2. Create a Branch

Work for the next release lives on `release/vX.Y.Z`, named after the version it will publish
(see [docs/RELEASING.md](docs/RELEASING.md)). Branch from it, and open your pull request against it:

```bash
git fetch origin
git checkout -b my-feature origin/release/v3.6.0   # the current release/v* branch
```

If there is no open `release/v*` branch, branch from `master` and open the PR against `master`;
a maintainer will retarget it.

### 3. Make Changes

* Follow Go best practices (`go fmt`, `go vet`)
* Keep the CLI flow simple and consistent
* If you add a feature, think about how it affects **developer flow**, not just functionality

### 4. Run Tests

```bash
go test ./...
go build -o /tmp/bffgen ./cmd/bffgen
BFFGEN=/tmp/bffgen test/e2e/run.sh nodejs-express /tmp/e2e   # also nodejs-fastify, python-fastapi
```

### 5. Commit & Push

```bash
git commit -m "feat: add my feature"
git push origin my-feature
```

### 6. Open a Pull Request

* Be clear about **what problem your PR solves**
* Include before/after flow notes if relevant
* Keep PRs small and focused

---

## 📋 What You Can Work On

* 🛠️ **Features** – Extend templates, add frameworks, improve scaffolding
* 🐞 **Bugs** – File and fix issues
* 📖 **Docs** – Improve explanations, examples, or tutorials
* 🔄 **Flows** – Suggest or test ways to make the CLI smoother
* 🧪 **DX Research** – Share your insights on how devs use CLI tools

---

## 🌱 Beginner-Friendly Contributions

Not sure where to start? Look for issues tagged:

* `good first issue`
* `help wanted`

These are perfect entry points for new contributors.

---

## 🧭 Community Guidelines

* Be kind and constructive
* Focus on improving **flows** for real developers
* Keep discussions transparent in GitHub Issues/PRs

---

## 📣 Recognition

All contributors will be listed in the README and thanked in release notes.

---

## 💡 Feedback & Ideas

If you’ve used `bffgen` and noticed a flow that felt rough or confusing, please [open an issue](https://github.com/RichGod93/bffgen/issues) or share your insights in the [survey](https://github.com/RichGod93/bffgen/discussions).

Your feedback is as valuable as code.

---

🙌 Thanks again for contributing — let’s make **bffgen** not just feature-rich, but *flow-first*.
