import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { Routes, Route, Link, useNavigate, useLocation } from 'react-router-dom';
import { getTheme } from '../../shared/styles/theme';
import AuthService from '../../shared/services/AuthService';
import CashierManagement from './components/CashierManagement';
import Dashboard from './components/Dashboard';
import WashBoxManagement from './components/WashBoxManagement';
import SessionManagement from './components/SessionManagement';
import QueueStatus from './components/QueueStatus';
import UserManagement from './components/UserManagement';
import PaymentManagement from './components/PaymentManagement';
import SettingsManagement from './components/SettingsManagement';


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
  
  &.active {
    color: ${props => props.theme.primaryColor};
    border-bottom: 2px solid ${props => props.theme.primaryColor};
  }
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
      
      <Navigation theme={theme}>
        <NavList>
          <NavItem>
            <NavLink to="/admin" theme={theme}>
              Панель управления
            </NavLink>
          </NavItem>
          <NavItem>
            <NavLink to="/admin/washboxes" theme={theme}>
              Боксы мойки
            </NavLink>
          </NavItem>
          <NavItem>
            <NavLink to="/admin/sessions" theme={theme}>
              Сессии мойки
            </NavLink>
          </NavItem>
          <NavItem>
            <NavLink to="/admin/queue" theme={theme}>
              Очередь
            </NavLink>
          </NavItem>
          <NavItem>
            <NavLink to="/admin/users" theme={theme}>
              Пользователи
            </NavLink>
          </NavItem>
          <NavItem>
            <NavLink to="/admin/cashiers" theme={theme}>
              Управление кассирами
            </NavLink>
          </NavItem>
          <NavItem>
            <NavLink to="/admin/payments" theme={theme}>
              Платежи
            </NavLink>
          </NavItem>
          <NavItem>
            <NavLink to="/admin/settings" theme={theme}>
              Настройки
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
          <Route path="/payments" element={<PaymentManagement />} />
          <Route path="/settings" element={<SettingsManagement />} />
  
        </Routes>
      </Content>
    </AdminContainer>
  );
};

export default AdminApp;
