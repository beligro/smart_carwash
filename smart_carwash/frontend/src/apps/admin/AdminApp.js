import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { Routes, Route, Link, Navigate, useNavigate, useLocation } from 'react-router-dom';
import { getTheme } from '../../shared/styles/theme';
import AuthService from '../../shared/services/AuthService';
import CashierManagement from './components/CashierManagement';
import CleanerManagement from './components/CleanerManagement';
import Dashboard from './components/Dashboard';
import WashBoxManagement from './components/WashBoxManagement';
import SessionManagement from './components/SessionManagement';
import QueueStatus from './components/QueueStatus';
import UserManagement from './components/UserManagement';
import PaymentManagement from './components/PaymentManagement';
import SettingsManagement from './components/SettingsManagement';
import ModbusDashboard from './components/ModbusDashboard';
import CleaningLogsManagement from './components/CleaningLogsManagement';
import WashboxChangeLogs from './components/WashboxChangeLogs';
import BoxMaintenanceManagement from './components/BoxMaintenanceManagement';
import ServiceTicketsManagement from './components/ServiceTicketsManagement';
import AdminManagement from './components/AdminManagement';
import MyWash from './components/MyWash';
import PersonalWashReport from './components/PersonalWashReport';


const AdminContainer = styled.div`
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: ${props => props.theme.backgroundColor};
  color: ${props => props.theme.textColor};
`;

const Header = styled.header`
  background-color: ${props => props.theme.cardBackground};
  padding: 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
`;

const Title = styled.h1`
  margin: 0;
  font-size: 1.5rem;
`;

const Navigation = styled.nav`
  background-color: ${props => props.theme.cardBackground};
  padding: 10px 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
`;

const NavList = styled.ul`
  display: flex;
  list-style: none;
  margin: 0;
  padding: 0;
  flex-wrap: wrap;
`;

const NavItem = styled.li`
  margin-right: 20px;
  
  &:last-child {
    margin-right: 0;
  }
`;

const NavLink = styled(Link)`
  color: ${props => props.theme.textColor};
  text-decoration: none;
  font-weight: 500;
  padding: 5px 0;
  
  &:hover {
    color: ${props => props.theme.primaryColor};
  }
  
  ${props => props.isActive && `
    color: ${props.theme.primaryColor};
    border-bottom: 2px solid ${props.theme.primaryColor};
  `}
`;

const LogoutButton = styled.button`
  background: none;
  border: none;
  color: ${props => props.theme.textColor};
  cursor: pointer;
  font-size: 1rem;
  
  &:hover {
    color: ${props => props.theme.primaryColor};
  }
`;

const UserInfo = styled.div`
  display: flex;
  align-items: center;
`;

const Username = styled.span`
  margin-right: 15px;
  font-weight: 500;
`;

const Content = styled.main`
  flex: 1;
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
`;

// Мобильная навигация
const MobileMenuButton = styled.button`
  display: none;
  background: none;
  border: none;
  color: ${props => props.theme.textColor};
  font-size: 1.5rem;
  cursor: pointer;
  padding: 8px;
  
  @media (max-width: 768px) {
    display: block;
  }
`;

const MobileMenu = styled.div`
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 1000;
  
  @media (max-width: 768px) {
    display: ${props => props.isOpen ? 'block' : 'none'};
  }
`;

const MobileMenuContent = styled.div`
  position: absolute;
  top: 0;
  left: 0;
  width: 280px;
  height: 100%;
  background-color: ${props => props.theme.cardBackground};
  padding: 20px;
  overflow-y: auto;
  box-shadow: 2px 0 10px rgba(0, 0, 0, 0.1);
`;

const MobileMenuHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 1px solid ${props => props.theme.borderColor};
`;

const MobileMenuTitle = styled.h3`
  margin: 0;
  color: ${props => props.theme.textColor};
`;

const MobileMenuCloseButton = styled.button`
  background: none;
  border: none;
  color: ${props => props.theme.textColor};
  font-size: 1.5rem;
  cursor: pointer;
  padding: 8px;
`;

const MobileNavList = styled.ul`
  list-style: none;
  margin: 0;
  padding: 0;
`;

const MobileNavItem = styled.li`
  margin-bottom: 10px;
`;

const MobileNavLink = styled(Link)`
  display: block;
  color: ${props => props.theme.textColor};
  text-decoration: none;
  font-weight: 500;
  padding: 12px 16px;
  border-radius: 6px;
  transition: background-color 0.2s;
  
  &:hover {
    background-color: ${props => props.theme.backgroundColor};
    color: ${props => props.theme.textColor};
  }
  
  ${props => props.isActive && `
    background-color: ${props.theme.primaryColor};
    color: white;
  `}
`;

/**
 * Приложение администратора
 * @returns {React.ReactNode} - Приложение администратора
 */
const AdminApp = () => {
  const theme = getTheme('light');
  const navigate = useNavigate();
  const location = useLocation();
  const [user, setUser] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const role = AuthService.getRole();
  const isLimitedAdmin = role === 'limited_admin';
  const isSuper = role === 'super_admin';
  const can = (section) => AuthService.canAccessSection(section);
  // Первый доступный раздел — куда отправлять админа без доступа к «Панели управления»
  const SECTION_PATHS = [
    ['dashboard', '/admin'],
    ['sessions', '/admin/sessions'],
    ['queue', '/admin/queue'],
    ['maintenance', '/admin/maintenance'],
    ['service-tickets', '/admin/service-tickets'],
    ['my-wash', '/admin/my-wash'],
    ['personal-wash-report', '/admin/personal-wash-report'],
    ['washboxes', '/admin/washboxes'],
    ['users', '/admin/users'],
    ['cashiers', '/admin/cashiers'],
    ['cleaners', '/admin/cleaners'],
    ['cleaning-logs', '/admin/cleaning-logs'],
    ['washbox-change-logs', '/admin/washbox-change-logs'],
    ['payments', '/admin/payments'],
    ['settings', '/admin/settings'],
    ['modbus-dashboard', '/admin/modbus-dashboard'],
  ];
  const firstAllowedPath = (SECTION_PATHS.find(([s]) => can(s)) || ['', '/admin'])[1];
  
  useEffect(() => {
    // Проверяем авторизацию при загрузке компонента
    const checkAuth = () => {
      const currentUser = AuthService.getCurrentUser();
      const isAuthenticated = AuthService.isAuthenticated();
      const isAdmin = AuthService.isAdmin();
      
      // Если пользователь не авторизован, перенаправляем на страницу входа
      if (!isAuthenticated) {
        navigate('/admin/login', { replace: true });
        return;
      }
      
      // Если пользователь не администратор, перенаправляем на страницу кассира
      if (!isAdmin) {
        navigate('/cashier', { replace: true });
        return;
      }
      
      // Если все проверки пройдены, устанавливаем пользователя
      setUser(currentUser);
      setIsLoading(false);
    };
    
    checkAuth();
  }, [navigate]);
  
  // Обработчик выхода из системы
  const handleLogout = async () => {
    await AuthService.logout();
    navigate('/admin/login');
  };
  
  // Если идет загрузка, показываем пустой контент
  if (isLoading) {
    return null;
  }
  
  return (
    <AdminContainer theme={theme}>
      <Header theme={theme}>
        <MobileMenuButton 
          theme={theme} 
          onClick={() => setIsMobileMenuOpen(true)}
        >
          ☰
        </MobileMenuButton>
        <Title>Интерфейс администратора</Title>
        {user && (
          <UserInfo>
            <Username>{user.username}</Username>
            <LogoutButton onClick={handleLogout} theme={theme}>
              Выйти
            </LogoutButton>
          </UserInfo>
        )}
      </Header>
      
      <Navigation theme={theme} className="desktop-nav">
        <NavList>
          {can('dashboard') && (
            <NavItem>
              <NavLink to="/admin" theme={theme} isActive={location.pathname === '/admin'}>
                Панель управления
              </NavLink>
            </NavItem>
          )}
          {can('washboxes') && (
            <NavItem>
              <NavLink to="/admin/washboxes" theme={theme} isActive={location.pathname === '/admin/washboxes'}>
                Боксы мойки
              </NavLink>
            </NavItem>
          )}
          {can('my-wash') && (
            <NavItem>
              <NavLink to="/admin/my-wash" theme={theme} isActive={location.pathname === '/admin/my-wash'}>
                Моя мойка
              </NavLink>
            </NavItem>
          )}
          {can('personal-wash-report') && (
            <NavItem>
              <NavLink to="/admin/personal-wash-report" theme={theme} isActive={location.pathname === '/admin/personal-wash-report'}>
                Личные мойки (отчёт)
              </NavLink>
            </NavItem>
          )}
          {can('maintenance') && (
            <NavItem>
              <NavLink to="/admin/maintenance" theme={theme} isActive={location.pathname === '/admin/maintenance'}>
                ТО аппаратов
              </NavLink>
            </NavItem>
          )}
          {can('service-tickets') && (
            <NavItem>
              <NavLink to="/admin/service-tickets" theme={theme} isActive={location.pathname === '/admin/service-tickets'}>
                Сервисные наряды
              </NavLink>
            </NavItem>
          )}
          {can('sessions') && (
            <NavItem>
              <NavLink to="/admin/sessions" theme={theme} isActive={location.pathname === '/admin/sessions'}>
                Сессии мойки
              </NavLink>
            </NavItem>
          )}
          {can('queue') && (
            <NavItem>
              <NavLink to="/admin/queue" theme={theme} isActive={location.pathname === '/admin/queue'}>
                Очередь
              </NavLink>
            </NavItem>
          )}
          {can('users') && (
            <NavItem>
              <NavLink to="/admin/users" theme={theme} isActive={location.pathname === '/admin/users'}>
                Клиенты
              </NavLink>
            </NavItem>
          )}
          {can('cashiers') && (
            <NavItem>
              <NavLink to="/admin/cashiers" theme={theme} isActive={location.pathname === '/admin/cashiers'}>
                Управление кассирами
              </NavLink>
            </NavItem>
          )}
          {can('cleaners') && (
            <NavItem>
              <NavLink to="/admin/cleaners" theme={theme} isActive={location.pathname === '/admin/cleaners'}>
                Управление уборщиками
              </NavLink>
            </NavItem>
          )}
          {can('cleaning-logs') && (
            <NavItem>
              <NavLink to="/admin/cleaning-logs" theme={theme} isActive={location.pathname === '/admin/cleaning-logs'}>
                Логи уборки
              </NavLink>
            </NavItem>
          )}
          {can('washbox-change-logs') && (
            <NavItem>
              <NavLink to="/admin/washbox-change-logs" theme={theme} isActive={location.pathname === '/admin/washbox-change-logs'}>
                История боксов
              </NavLink>
            </NavItem>
          )}
          {can('payments') && (
            <NavItem>
              <NavLink to="/admin/payments" theme={theme} isActive={location.pathname === '/admin/payments'}>
                Платежи
              </NavLink>
            </NavItem>
          )}
          {can('settings') && (
            <NavItem>
              <NavLink to="/admin/settings" theme={theme} isActive={location.pathname === '/admin/settings'}>
                Настройки
              </NavLink>
            </NavItem>
          )}
          {can('modbus-dashboard') && (
            <NavItem>
              <NavLink to="/admin/modbus-dashboard" theme={theme} isActive={location.pathname === '/admin/modbus-dashboard'}>
                Modbus мониторинг
              </NavLink>
            </NavItem>
          )}
          {isSuper && (
            <NavItem>
              <NavLink to="/admin/admins" theme={theme} isActive={location.pathname === '/admin/admins'}>
                Управление администраторами
              </NavLink>
            </NavItem>
          )}
        </NavList>
      </Navigation>
      
      <Content>
        <Routes>
          <Route path="/" element={can('dashboard') ? <Dashboard /> : <Navigate to={firstAllowedPath} replace />} />
          {can('washboxes') && <Route path="/washboxes" element={<WashBoxManagement />} />}
          {can('my-wash') && <Route path="/my-wash" element={<MyWash />} />}
          {can('personal-wash-report') && <Route path="/personal-wash-report" element={<PersonalWashReport />} />}
          {can('maintenance') && <Route path="/maintenance" element={<BoxMaintenanceManagement />} />}
          {can('service-tickets') && <Route path="/service-tickets" element={<ServiceTicketsManagement />} />}
          {can('sessions') && <Route path="/sessions" element={<SessionManagement />} />}
          {can('queue') && <Route path="/queue" element={<QueueStatus />} />}
          {can('users') && <Route path="/users" element={<UserManagement />} />}
          {can('cashiers') && <Route path="/cashiers" element={<CashierManagement />} />}
          {can('cleaners') && <Route path="/cleaners" element={<CleanerManagement />} />}
          {can('cleaning-logs') && <Route path="/cleaning-logs" element={<CleaningLogsManagement />} />}
          {can('washbox-change-logs') && !isLimitedAdmin && <Route path="/washbox-change-logs" element={<WashboxChangeLogs theme={theme} />} />}
          {can('payments') && <Route path="/payments" element={<PaymentManagement />} />}
          {can('settings') && <Route path="/settings" element={<SettingsManagement />} />}
          {can('modbus-dashboard') && <Route path="/modbus-dashboard" element={<ModbusDashboard />} />}
          {isSuper && <Route path="/admins" element={<AdminManagement />} />}
          <Route path="*" element={<Navigate to={firstAllowedPath} replace />} />
        </Routes>
      </Content>

      {/* Мобильное меню */}
      <MobileMenu isOpen={isMobileMenuOpen} onClick={() => setIsMobileMenuOpen(false)}>
        <MobileMenuContent theme={theme} onClick={(e) => e.stopPropagation()}>
          <MobileMenuHeader theme={theme}>
            <MobileMenuTitle theme={theme}>Меню</MobileMenuTitle>
            <MobileMenuCloseButton 
              theme={theme} 
              onClick={() => setIsMobileMenuOpen(false)}
            >
              ×
            </MobileMenuCloseButton>
          </MobileMenuHeader>
          
          <MobileNavList>
            {can('dashboard') && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin" 
                  theme={theme} 
                  isActive={location.pathname === '/admin'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  Панель управления
                </MobileNavLink>
              </MobileNavItem>
            )}
            {can('washboxes') && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin/washboxes" 
                  theme={theme} 
                  isActive={location.pathname === '/admin/washboxes'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  Боксы мойки
                </MobileNavLink>
              </MobileNavItem>
            )}
            {can('my-wash') && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin/my-wash" 
                  theme={theme} 
                  isActive={location.pathname === '/admin/my-wash'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  Моя мойка
                </MobileNavLink>
              </MobileNavItem>
            )}
            {can('personal-wash-report') && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin/personal-wash-report" 
                  theme={theme} 
                  isActive={location.pathname === '/admin/personal-wash-report'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  Личные мойки (отчёт)
                </MobileNavLink>
              </MobileNavItem>
            )}
            {can('maintenance') && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin/maintenance" 
                  theme={theme} 
                  isActive={location.pathname === '/admin/maintenance'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  ТО аппаратов
                </MobileNavLink>
              </MobileNavItem>
            )}
            {can('service-tickets') && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin/service-tickets" 
                  theme={theme} 
                  isActive={location.pathname === '/admin/service-tickets'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  Сервисные наряды
                </MobileNavLink>
              </MobileNavItem>
            )}
            {can('sessions') && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin/sessions" 
                  theme={theme} 
                  isActive={location.pathname === '/admin/sessions'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  Сессии мойки
                </MobileNavLink>
              </MobileNavItem>
            )}
            {can('queue') && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin/queue" 
                  theme={theme} 
                  isActive={location.pathname === '/admin/queue'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  Очередь
                </MobileNavLink>
              </MobileNavItem>
            )}
            {can('users') && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin/users" 
                  theme={theme} 
                  isActive={location.pathname === '/admin/users'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  Клиенты
                </MobileNavLink>
              </MobileNavItem>
            )}
            {can('cashiers') && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin/cashiers" 
                  theme={theme} 
                  isActive={location.pathname === '/admin/cashiers'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  Управление кассирами
                </MobileNavLink>
              </MobileNavItem>
            )}
            {can('cleaners') && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin/cleaners" 
                  theme={theme} 
                  isActive={location.pathname === '/admin/cleaners'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  Управление уборщиками
                </MobileNavLink>
              </MobileNavItem>
            )}
            {can('cleaning-logs') && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin/cleaning-logs" 
                  theme={theme} 
                  isActive={location.pathname === '/admin/cleaning-logs'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  Логи уборки
                </MobileNavLink>
              </MobileNavItem>
            )}
            {can('washbox-change-logs') && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin/washbox-change-logs" 
                  theme={theme} 
                  isActive={location.pathname === '/admin/washbox-change-logs'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  История боксов
                </MobileNavLink>
              </MobileNavItem>
            )}
            {can('payments') && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin/payments" 
                  theme={theme} 
                  isActive={location.pathname === '/admin/payments'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  Платежи
                </MobileNavLink>
              </MobileNavItem>
            )}
            {can('settings') && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin/settings" 
                  theme={theme} 
                  isActive={location.pathname === '/admin/settings'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  Настройки
                </MobileNavLink>
              </MobileNavItem>
            )}
            {can('modbus-dashboard') && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin/modbus-dashboard" 
                  theme={theme} 
                  isActive={location.pathname === '/admin/modbus-dashboard'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  Modbus мониторинг
                </MobileNavLink>
              </MobileNavItem>
            )}
            {isSuper && (
              <MobileNavItem>
                <MobileNavLink 
                  to="/admin/admins" 
                  theme={theme} 
                  isActive={location.pathname === '/admin/admins'}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  Управление администраторами
                </MobileNavLink>
              </MobileNavItem>
            )}
          </MobileNavList>
        </MobileMenuContent>
      </MobileMenu>
    </AdminContainer>
  );
};

export default AdminApp;
