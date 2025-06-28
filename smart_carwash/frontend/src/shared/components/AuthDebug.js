import React from 'react';
import styled from 'styled-components';
import AuthService from '../services/AuthService';

const DebugContainer = styled.div`
  position: fixed;
  top: 10px;
  right: 10px;
  background-color: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 10px;
  border-radius: 5px;
  font-size: 12px;
  z-index: 1000;
  max-width: 300px;
`;

const DebugItem = styled.div`
  margin-bottom: 5px;
`;

const AuthDebug = () => {
  const isAuthenticated = AuthService.isAuthenticated();
  const isAdmin = AuthService.isAdmin();
  const currentUser = AuthService.getCurrentUser();
  const token = localStorage.getItem('token');
  const expiresAt = localStorage.getItem('expiresAt');

  // Показываем только в режиме разработки
  if (process.env.NODE_ENV !== 'development') {
    return null;
  }

  return (
    <DebugContainer>
      <DebugItem>
        <strong>Auth Status:</strong> {isAuthenticated ? '✅ Authenticated' : '❌ Not Authenticated'}
      </DebugItem>
      <DebugItem>
        <strong>Role:</strong> {isAdmin ? '👑 Admin' : '👤 Cashier'}
      </DebugItem>
      <DebugItem>
        <strong>User:</strong> {currentUser ? currentUser.username : 'None'}
      </DebugItem>
      <DebugItem>
        <strong>Token:</strong> {token ? `${token.substring(0, 20)}...` : 'None'}
      </DebugItem>
      <DebugItem>
        <strong>Expires:</strong> {expiresAt ? new Date(expiresAt).toLocaleString() : 'None'}
      </DebugItem>
      <DebugItem>
        <strong>Current Path:</strong> {window.location.pathname}
      </DebugItem>
    </DebugContainer>
  );
};

export default AuthDebug; 