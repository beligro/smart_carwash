import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { Routes, Route, Link, useNavigate, useLocation } from 'react-router-dom';
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
          <NavItem>
            <NavLink to="/admin" theme={theme} isActive={location.pathname === '/admin'}>
              Панель управления
            </NavLink>
          </NavItem>
          <NavItem>
            <NavLink to="/admin/washboxes" theme={theme} isActive={location.pathname === '/admin/washboxes'}>
              Боксы мойки
            </NavLink>
          </NavItem>
          <NavItem>
            <NavLink to="/admin/sessions" theme={theme} isActive={location.pathname === '/admin/sessions'}>
              Сессии мойки
            </NavLink>
          </NavItem>
          <NavItem>
            <NavLink to="/admin/queue" theme={theme} isActive={location.pathname === '/admin/queue'}>
              Очередь
            </NavLink>
          </NavItem>
          <NavItem>
            <NavLink to="/admin/users" theme={theme} isActive={location.pathname === '/admin/users'}>
              Клиенты
            </NavLink>
          </NavItem>
          <NavItem>
            <NavLink to="/admin/cashiers" theme={theme} isActive={location.pathname === '/admin/cashiers'}>
              Управление кассирами
            </NavLink>
          </NavItem>
          <NavItem>
            <NavLink to="/admin/cleaners" theme={theme} isActive={location.pathname === '/admin/cleaners'}>
              Управление уборщиками
            </NavLink>
          </NavItem>
          <NavItem>
            <NavLink to="/admin/cleaning-logs" theme={theme} isActive={location.pathname === '/admin/cleaning-logs'}>
              Логи уборки
            </NavLink>
          </NavItem>
          {!isLimitedAdmin && (
            <NavItem>
              <NavLink to="/admin/washbox-change-logs" theme={theme} isActive={location.pathname === '/admin/washbox-change-logs'}>
                История боксов
              </NavLink>
            </NavItem>
          )}
          <NavItem>
            <NavLink to="/admin/payments" theme={theme} isActive={location.pathname === '/admin/payments'}>
              Платежи
            </NavLink>
          </NavItem>
          <NavItem>
            <NavLink to="/admin/settings" theme={theme} isActive={location.pathname === '/admin/settings'}>
              Настройки
            </NavLink>
          </NavItem>
          <NavItem>
            <NavLink to="/admin/modbus-dashboard" theme={theme} isActive={location.pathname === '/admin/modbus-dashboard'}>
              Modbus мониторинг
            </NavLink>
          </NavItem>
        </NavList>
      </Navigation>
      
      <Content>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/washboxes" element={<WashBoxManagement />} />
          <Route path="/sessions" element={<SessionManagement />} />
          <Route path="/queue" element={<QueueStatus />} />
          <Route path="/users" element={<UserManagement />} />
          <Route path="/cashiers" element={<CashierManagement />} />
          <Route path="/cleaners" element={<CleanerManagement />} />
          <Route path="/cleaning-logs" element={<CleaningLogsManagement />} />
          {!isLimitedAdmin && <Route path="/washbox-change-logs" element={<WashboxChangeLogs theme={theme} />} />}
          <Route path="/payments" element={<PaymentManagement />} />
          <Route path="/settings" element={<SettingsManagement />} />
          <Route path="/modbus-dashboard" element={<ModbusDashboard />} />
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
          </MobileNavList>
        </MobileMenuContent>
      </MobileMenu>
    </AdminContainer>
  );
};

export default AdminApp;
