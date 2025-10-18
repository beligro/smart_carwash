import React from 'react';
import styled from 'styled-components';

// Стили для мобильных карточек
const MobileCard = styled.div`
  display: none;
  background-color: ${props => props.theme.cardBackground};
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  border-left: 4px solid ${props => props.borderColor || '#6c757d'};
  
  @media (max-width: 768px) {
    display: block;
  }
`;

const MobileCardHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
`;

const MobileCardTitle = styled.div`
  font-weight: 600;
  font-size: 1rem;
  color: ${props => props.theme.textColor};
`;

const MobileCardStatus = styled.div`
  font-size: 0.8rem;
`;

const MobileCardDetails = styled.div`
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-bottom: 12px;
`;

const MobileCardDetail = styled.div`
  display: flex;
  flex-direction: column;
`;

const MobileCardLabel = styled.span`
  font-size: 0.7rem;
  color: ${props => props.theme.textColorSecondary};
  margin-bottom: 2px;
`;

const MobileCardValue = styled.span`
  font-size: 0.8rem;
  color: ${props => props.theme.textColor};
  font-weight: 500;
`;

const MobileCardActions = styled.div`
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
`;

const MobileActionButton = styled.button`
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  min-height: 32px;
  min-width: 32px;

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  &.primary {
    background-color: ${props => props.theme.primaryColor};
    color: white;
    
    &:hover:not(:disabled) {
      opacity: 0.9;
    }
  }

  &.secondary {
    background-color: #6c757d;
    color: white;
    
    &:hover:not(:disabled) {
      background-color: #5a6268;
    }
  }

  &.danger {
    background-color: #dc3545;
    color: white;
    
    &:hover:not(:disabled) {
      background-color: #c82333;
    }
  }

  &.success {
    background-color: #28a745;
    color: white;
    
    &:hover:not(:disabled) {
      background-color: #218838;
    }
  }
`;

/**
 * Универсальный компонент для мобильных таблиц
 * @param {Object} props - Пропсы компонента
 * @param {Array} props.data - Массив данных для отображения
 * @param {Array} props.columns - Конфигурация колонок
 * @param {Function} props.getBorderColor - Функция для получения цвета границы карточки
 * @param {Function} props.renderActions - Функция для рендеринга действий
 * @param {Object} props.theme - Тема
 * @returns {React.ReactNode} - Компонент мобильных карточек
 */
const MobileTable = ({ 
  data, 
  columns, 
  getBorderColor, 
  renderActions, 
  theme,
  titleField = 'id',
  statusField = 'status'
}) => {
  if (!data || data.length === 0) {
    return null;
  }

  return (
    <div className="mobile-card">
      {data.map((item, index) => {
        const borderColor = getBorderColor ? getBorderColor(item) : '#6c757d';
        const title = item[titleField] || `Элемент ${index + 1}`;
        const status = item[statusField];

        return (
          <MobileCard key={item.id || index} theme={theme} borderColor={borderColor}>
            <MobileCardHeader>
              <MobileCardTitle theme={theme}>
                {typeof title === 'string' && title.length > 20 
                  ? `${title.substring(0, 20)}...` 
                  : title}
              </MobileCardTitle>
              {status && (
                <MobileCardStatus>
                  {status}
                </MobileCardStatus>
              )}
            </MobileCardHeader>
            
            <MobileCardDetails>
              {columns.map((column, colIndex) => {
                if (column.hideOnMobile) return null;
                
                const value = column.accessor ? column.accessor(item) : item[column.key];
                const displayValue = column.render ? column.render(value, item) : value;

                return (
                  <MobileCardDetail key={colIndex}>
                    <MobileCardLabel theme={theme}>
                      {column.label}
                    </MobileCardLabel>
                    <MobileCardValue theme={theme}>
                      {displayValue}
                    </MobileCardValue>
                  </MobileCardDetail>
                );
              })}
            </MobileCardDetails>
            
            {renderActions && (
              <MobileCardActions>
                {renderActions(item)}
              </MobileCardActions>
            )}
          </MobileCard>
        );
      })}
    </div>
  );
};

export default MobileTable;
