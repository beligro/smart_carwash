import React from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../styles/theme';

const ModalOverlay = styled.div`
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
`;

const ModalContent = styled.div`
  background-color: white;
  border-radius: 8px;
  padding: 24px;
  max-width: 500px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  border: 1px solid #e0e0e0;
`;

const ModalHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
`;

const ModalTitle = styled.h3`
  margin: 0;
  color: #333;
  font-size: 1.25rem;
  font-weight: 600;
`;

const CloseButton = styled.button`
  background: none;
  border: none;
  font-size: 1.5rem;
  color: #666;
  cursor: pointer;
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;

  &:hover {
    color: #333;
  }
`;

const WarningIcon = styled.div`
  background-color: #ff9800;
  color: white;
  border-radius: 50%;
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  margin: 0 auto 16px;
`;

const WarningMessage = styled.div`
  background-color: #fff3cd;
  border: 1px solid #ffeaa7;
  border-radius: 6px;
  padding: 16px;
  margin-bottom: 20px;
`;

const WarningText = styled.p`
  margin: 0;
  color: #856404;
  font-size: 0.9rem;
  line-height: 1.4;
`;

const ConsequencesList = styled.ul`
  margin: 0;
  padding-left: 20px;
  color: #333;
  font-size: 0.9rem;
  line-height: 1.5;
`;

const ConsequenceItem = styled.li`
  margin-bottom: 8px;
`;

const ButtonContainer = styled.div`
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
`;

const Button = styled.button`
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;

  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
`;

const CancelButton = styled(Button)`
  background-color: ${props => props.theme.textColorSecondary};
  color: white;

  &:hover:not(:disabled) {
    background-color: ${props => props.theme.textColor};
  }
`;

const ReassignButton = styled(Button)`
  background-color: #ff9800;
  color: white;

  &:hover:not(:disabled) {
    background-color: #f57c00;
  }
`;

const ReassignSessionModal = ({ 
  isOpen, 
  onClose, 
  onConfirm, 
  sessionId, 
  serviceType,
  isLoading = false 
}) => {
  if (!isOpen) return null;

  const getServiceText = (type) => {
    switch (type) {
      case 'wash': return 'мойки';
      case 'air_dry': return 'обдува';
      case 'vacuum': return 'пылесоса';
      default: return 'услуги';
    }
  };

  const handleConfirm = () => {
    onConfirm(sessionId);
  };

  return (
    <ModalOverlay onClick={onClose}>
      <ModalContent onClick={(e) => e.stopPropagation()}>
        <ModalHeader>
          <ModalTitle>Переназначение сессии</ModalTitle>
          <CloseButton onClick={onClose}>×</CloseButton>
        </ModalHeader>

        <WarningIcon>⚠️</WarningIcon>

        <WarningMessage>
          <WarningText>
            <strong>Внимание!</strong> Вы собираетесь переназначить сессию {getServiceText(serviceType)} на другой бокс.
          </WarningText>
        </WarningMessage>

        <div>
          <p style={{ margin: '0 0 12px 0', color: '#333', fontWeight: '500' }}>
            Это действие приведет к следующим последствиям:
          </p>
          
          <ConsequencesList>
            <ConsequenceItem>
              <strong>Бокс будет переведен в статус "На обслуживании"</strong>
            </ConsequenceItem>
            <ConsequenceItem>
              <strong>Свет и химия будут выключены</strong> в старом боксе
            </ConsequenceItem>
            <ConsequenceItem>
              <strong>Сессия вернется в очередь</strong> и будет назначена на новый бокс
            </ConsequenceItem>
            <ConsequenceItem>
              <strong>Таймер сессии сбросится</strong> - время начнет отсчитываться заново
            </ConsequenceItem>
            <ConsequenceItem>
              <strong>Пользователь получит уведомление</strong> о переназначении и должен будет переехать к новому боксу
            </ConsequenceItem>
          </ConsequencesList>
        </div>

        <ButtonContainer>
          <CancelButton onClick={onClose} disabled={isLoading}>
            Отмена
          </CancelButton>
          <ReassignButton onClick={handleConfirm} disabled={isLoading}>
            {isLoading ? 'Переназначаем...' : 'Переназначить'}
          </ReassignButton>
        </ButtonContainer>
      </ModalContent>
    </ModalOverlay>
  );
};

export default ReassignSessionModal;
