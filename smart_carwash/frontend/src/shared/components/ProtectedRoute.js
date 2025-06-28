import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import AuthService from '../services/AuthService';

/**
 * Компонент для защиты маршрутов, требующих авторизации
 * @param {Object} props - Свойства компонента
 * @param {React.ReactNode} props.children - Дочерние компоненты
 * @param {boolean} props.requireAdmin - Требуется ли доступ администратора
 * @returns {React.ReactNode} - Защищенный маршрут или перенаправление
 */
const ProtectedRoute = ({ children, requireAdmin = false }) => {
  const location = useLocation();
  const isAuthenticated = AuthService.isAuthenticated();
  const isAdmin = AuthService.isAdmin();

  // Если пользователь не авторизован, перенаправляем на страницу входа
  if (!isAuthenticated) {
    // Определяем, на какую страницу входа перенаправить
    const loginPath = requireAdmin ? '/admin/login' : '/cashier/login';
    
    // Сохраняем текущий путь для перенаправления после входа
    return <Navigate to={loginPath} state={{ from: location }} replace />;
  }

  // Если требуется доступ администратора, но пользователь не администратор
  if (requireAdmin && !isAdmin) {
    return <Navigate to="/cashier" replace />;
  }

  // Если пользователь администратор, но пытается получить доступ к странице кассира
  if (!requireAdmin && isAdmin && location.pathname.startsWith('/cashier')) {
    return <Navigate to="/admin" replace />;
  }

  // Если кассир пытается получить доступ к странице администратора
  if (requireAdmin && !isAdmin && location.pathname.startsWith('/admin')) {
    return <Navigate to="/cashier" replace />;
  }

  // Если все проверки пройдены, отображаем защищенный контент
  return children;
};

export default ProtectedRoute;
