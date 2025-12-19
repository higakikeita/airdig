import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { Zap, Activity, Database } from 'lucide-react';

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const location = useLocation();

  const navItems = [
    { path: '/', label: 'Dashboard', icon: Activity },
    { path: '/resources', label: 'Resources', icon: Database },
  ];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100vh', overflow: 'hidden' }}>
      {/* Header - Compact */}
      <div style={{
        background: 'var(--aws-squid-ink)',
        borderBottom: '1px solid var(--aws-smile-orange)',
        padding: '6px var(--space-l)',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        height: '36px',
        zIndex: 1000,
      }}>
        <div style={{ fontSize: '14px', fontWeight: '700', color: 'white', letterSpacing: '0.3px' }}>
          <Zap size={18} style={{ display: 'inline', marginRight: '8px', color: 'var(--aws-smile-orange)', verticalAlign: 'middle' }} />
          AirDig <span style={{ color: 'var(--aws-smile-orange)' }}>DeepDrift</span>
        </div>
        <div style={{ display: 'flex', gap: '12px', alignItems: 'center' }}>
          <div style={{
            background: 'var(--color-background-button-primary-default)',
            color: 'white',
            padding: '2px 8px',
            borderRadius: '3px',
            fontSize: '11px',
            fontWeight: '500',
          }}>
            v0.1.0-alpha
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '12px' }}>
            <div style={{
              width: '6px',
              height: '6px',
              borderRadius: '50%',
              background: '#1d8102',
            }} />
            <span>Connected</span>
          </div>
        </div>
      </div>

      {/* Main Layout with Sidebar */}
      <div style={{ display: 'flex', height: 'calc(100vh - 36px)', overflow: 'hidden' }}>
        {/* Sidebar Navigation - Hide on Dashboard page */}
        {location.pathname !== '/' && (
          <div style={{
            width: '240px',
            background: 'var(--color-background-layout-panel-content)',
            borderRight: '1px solid var(--color-border-divider-default)',
            padding: 'var(--space-m)',
            display: 'flex',
            flexDirection: 'column',
            gap: 'var(--space-xxs)',
          }}>
            {navItems.map(({ path, label, icon: Icon }) => {
              const isActive = location.pathname === path;
              return (
                <Link
                  key={path}
                  to={path}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 'var(--space-s)',
                    padding: 'var(--space-s) var(--space-m)',
                    borderRadius: '8px',
                    textDecoration: 'none',
                    color: isActive ? 'var(--color-text-interactive-active)' : 'var(--color-text-body-default)',
                    background: isActive ? 'var(--color-background-button-primary-default)' : 'transparent',
                    fontWeight: isActive ? '600' : '400',
                    fontSize: 'var(--font-size-body-m)',
                    transition: 'all 0.2s',
                  }}
                  onMouseEnter={(e) => {
                    if (!isActive) {
                      e.currentTarget.style.background = 'var(--color-background-item-hover)';
                    }
                  }}
                  onMouseLeave={(e) => {
                    if (!isActive) {
                      e.currentTarget.style.background = 'transparent';
                    }
                  }}
                >
                  <Icon size={20} />
                  {label}
                </Link>
              );
            })}
          </div>
        )}

        {/* Main Content */}
        <div style={{ flex: 1, overflow: 'auto' }}>
          {children}
        </div>
      </div>
    </div>
  );
};

export default Layout;
