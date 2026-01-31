import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import { Search, UserCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useAuthStore } from '@/store/auth';
import { cn } from '@/lib/utils';

const navItems = [
  { label: 'Devices', to: '/app/devices' },
  { label: 'Telemetry', to: '/app/telemetry' },
  { label: 'Commands', to: '/app/commands' },
  { label: 'Settings', to: '/app/settings' },
];

export function AppLayout() {
  const navigate = useNavigate();
  const { logout, user } = useAuthStore();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="flex min-h-screen bg-background">
      <aside className="hidden w-64 flex-col border-r border-border/70 bg-card/60 p-6 md:flex">
        <div className="mb-10 text-lg font-semibold text-foreground">KONYX</div>
        <nav className="flex flex-1 flex-col gap-2">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                cn(
                  'rounded-lg px-4 py-2 text-sm text-muted-foreground transition hover:bg-muted/50 hover:text-foreground',
                  isActive && 'bg-muted text-foreground shadow-glow'
                )
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="text-xs text-muted-foreground">Premium operations suite</div>
      </aside>
      <div className="flex flex-1 flex-col">
        <header className="flex items-center justify-between gap-4 border-b border-border/70 bg-card/40 px-6 py-4">
          <div className="flex w-full max-w-md items-center gap-2 rounded-lg border border-border/70 bg-muted/30 px-3">
            <Search className="h-4 w-4 text-muted-foreground" />
            <Input placeholder="Search devices or telemetry..." className="border-0 bg-transparent px-0" />
          </div>
          <div className="flex items-center gap-3">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="flex items-center gap-2">
                  <UserCircle className="h-5 w-5" />
                  <span className="hidden text-sm md:inline">{user?.email ?? 'Operator'}</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuLabel>Account</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={() => navigate('/app/settings')}>Settings</DropdownMenuItem>
                <DropdownMenuItem onClick={handleLogout}>Logout</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </header>
        <main className="flex-1 bg-background px-6 py-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
