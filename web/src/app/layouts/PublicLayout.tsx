import { NavLink, Outlet } from 'react-router-dom';
import { Button } from '@/components/ui/button';

const navItems = [
  { label: 'О нас', to: '/about' },
  { label: 'Спецификация', to: '/spec' },
  { label: 'Как заказать', to: '/order' },
  { label: 'Поддержка', to: '/support' },
];

export function PublicLayout() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="sticky top-0 z-40 border-b border-border/60 bg-background/80 backdrop-blur">
        <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
          <NavLink to="/" className="text-lg font-semibold tracking-wide text-foreground">
            KONYX
          </NavLink>
          <nav className="hidden items-center gap-6 text-sm text-muted-foreground md:flex">
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) =>
                  `transition-colors hover:text-foreground ${isActive ? 'text-foreground' : ''}`
                }
              >
                {item.label}
              </NavLink>
            ))}
          </nav>
          <div className="flex items-center gap-3">
            <Button variant="ghost" asChild>
              <NavLink to="/login">Войти</NavLink>
            </Button>
            <Button asChild>
              <NavLink to="/register">Регистрация</NavLink>
            </Button>
          </div>
        </div>
      </header>
      <main className="relative">
        <div className="absolute inset-0 -z-10 bg-grid opacity-60" />
        <Outlet />
      </main>
      <footer className="border-t border-border/70 py-10">
        <div className="mx-auto flex max-w-6xl flex-col gap-3 px-6 text-sm text-muted-foreground md:flex-row md:items-center md:justify-between">
          <span>© {new Date().getFullYear()} KONYX</span>
          <span>Premium tech operations console.</span>
        </div>
      </footer>
    </div>
  );
}
