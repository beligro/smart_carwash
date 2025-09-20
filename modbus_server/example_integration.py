#!/usr/bin/env python3
"""
Пример интеграции с ПЛК210 веб-сервером
Для разработчика Артема

Этот файл показывает, как интегрировать ПЛК210 в ваш проект.
"""

import requests
import time
import json
from typing import Dict, List, Optional, Any

class PLC210Client:
    """
    Клиент для работы с ПЛК210 через веб-сервер
    
    Использование:
        plc = PLC210Client("http://localhost:5000")
        
        # Включить бит
        plc.set_bit(0, 5, True)
        
        # Получить состояние
        status = plc.get_status()
        
        # Включить все биты
        plc.turn_all(True)
    """
    
    def __init__(self, base_url: str = "http://localhost:5000"):
        """
        Инициализация клиента
        
        Args:
            base_url: Базовый URL веб-сервера ПЛК210
        """
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        self.session.timeout = 10
        
    def _make_request(self, method: str, endpoint: str, **kwargs) -> Optional[Dict]:
        """Выполнение HTTP запроса с обработкой ошибок"""
        try:
            url = f"{self.base_url}{endpoint}"
            response = self.session.request(method, url, **kwargs)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f"Ошибка запроса к {endpoint}: {e}")
            return None
        except json.JSONDecodeError as e:
            print(f"Ошибка парсинга JSON: {e}")
            return None
    
    def get_status(self) -> Optional[Dict[str, Any]]:
        """
        Получение полного состояния всех регистров
        
        Returns:
            Dict с состоянием всех регистров или None при ошибке
            
        Example:
            status = plc.get_status()
            if status and status['success']:
                registers = status['data']['registers']
                reg0_value = registers['register_0']['value']
                reg0_bits = registers['register_0']['bits']
        """
        return self._make_request('GET', '/api/status')
    
    def set_bit(self, register: int, bit: int, state: bool) -> bool:
        """
        Установка состояния конкретного бита
        
        Args:
            register: Номер регистра (0-5)
            bit: Номер бита (0-15, для register 5: 0-9)
            state: True для включения, False для выключения
            
        Returns:
            True если операция успешна, False при ошибке
            
        Example:
            # Включить бит 5 в регистре 0
            success = plc.set_bit(0, 5, True)
            
            # Выключить бит 10 в регистре 2
            success = plc.set_bit(2, 10, False)
        """
        data = {
            'register': register,
            'bit': bit,
            'state': state
        }
        
        result = self._make_request('POST', '/api/set_bit', json=data)
        return result and result.get('success', False)
    
    def get_bit(self, register: int, bit: int) -> Optional[bool]:
        """
        Получение состояния конкретного бита
        
        Args:
            register: Номер регистра (0-5)
            bit: Номер бита (0-15)
            
        Returns:
            True/False состояние бита или None при ошибке
            
        Example:
            state = plc.get_bit(0, 5)
            if state is not None:
                print(f"Бит 5 в регистре 0: {'ВКЛ' if state else 'ВЫКЛ'}")
        """
        params = {'register': register, 'bit': bit}
        result = self._make_request('GET', '/api/get_bit', params=params)
        
        if result and result.get('success'):
            return result.get('state')
        return None
    
    def read_register(self, register: int) -> Optional[Dict[str, Any]]:
        """
        Чтение полной информации о регистре
        
        Args:
            register: Номер регистра (0-5)
            
        Returns:
            Dict с информацией о регистре или None при ошибке
            
        Example:
            reg_info = plc.read_register(0)
            if reg_info and reg_info['success']:
                value = reg_info['value']      # Числовое значение
                hex_val = reg_info['hex']      # Hex представление
                bits = reg_info['bits']        # Массив битов [0,1,0,1,...]
        """
        params = {'register': register}
        return self._make_request('GET', '/api/read_register', params=params)
    
    def write_register(self, register: int, value: int) -> bool:
        """
        Запись значения в регистр целиком
        
        Args:
            register: Номер регистра (0-5)
            value: Значение (0-65535)
            
        Returns:
            True если операция успешна, False при ошибке
            
        Example:
            # Записать 0xAAAA (43690) в регистр 1
            success = plc.write_register(1, 43690)
            
            # Записать двоичное значение 1010101010101010
            success = plc.write_register(2, 0b1010101010101010)
        """
        data = {
            'register': register,
            'value': value
        }
        
        result = self._make_request('POST', '/api/write_register', json=data)
        return result and result.get('success', False)
    
    def turn_all(self, state: bool) -> bool:
        """
        Включение/выключение всех 90 битов
        
        Args:
            state: True для включения всех, False для выключения всех
            
        Returns:
            True если операция успешна, False при ошибке
            
        Example:
            # Включить все 90 битов
            success = plc.turn_all(True)
            
            # Выключить все 90 битов  
            success = plc.turn_all(False)
        """
        data = {'state': state}
        result = self._make_request('POST', '/api/turn_all', json=data)
        return result and result.get('success', False)
    
    def get_connection_info(self) -> Optional[Dict[str, Any]]:
        """
        Получение информации о подключении к ПЛК
        
        Returns:
            Dict с информацией о подключении или None при ошибке
            
        Example:
            conn_info = plc.get_connection_info()
            if conn_info and conn_info['success']:
                connected = conn_info['connection']['connected']
                host = conn_info['connection']['host']
                port = conn_info['connection']['port']
        """
        return self._make_request('GET', '/api/connection')
    
    def wait_for_connection(self, timeout: int = 30) -> bool:
        """
        Ожидание подключения к ПЛК
        
        Args:
            timeout: Максимальное время ожидания в секундах
            
        Returns:
            True если подключение установлено, False при таймауте
        """
        start_time = time.time()
        
        while time.time() - start_time < timeout:
            conn_info = self.get_connection_info()
            if conn_info and conn_info.get('success'):
                if conn_info['connection']['connected']:
                    return True
            
            time.sleep(1)
        
        return False

def example_basic_usage():
    """Базовый пример использования"""
    print("=== Базовый пример использования ПЛК210 ===")
    
    # Создаем клиент
    plc = PLC210Client("http://localhost:5000")
    
    # Проверяем подключение
    print("Проверка подключения к ПЛК...")
    if not plc.wait_for_connection(timeout=10):
        print("❌ Не удалось подключиться к ПЛК210")
        return
    
    print("✅ Подключение к ПЛК210 установлено")
    
    # Включаем бит 0 в регистре 0
    print("\n1. Включение бита 0 в регистре 0...")
    if plc.set_bit(0, 0, True):
        print("✅ Бит включен")
    else:
        print("❌ Ошибка включения бита")
    
    # Проверяем состояние бита
    print("\n2. Проверка состояния бита...")
    state = plc.get_bit(0, 0)
    if state is not None:
        print(f"✅ Состояние бита 0 в регистре 0: {'ВКЛ' if state else 'ВЫКЛ'}")
    else:
        print("❌ Ошибка чтения состояния бита")
    
    # Читаем весь регистр
    print("\n3. Чтение регистра 0...")
    reg_info = plc.read_register(0)
    if reg_info and reg_info.get('success'):
        print(f"✅ Значение регистра 0: {reg_info['value']} ({reg_info['hex']})")
        print(f"   Биты: {reg_info['bits']}")
    else:
        print("❌ Ошибка чтения регистра")

def example_pattern_control():
    """Пример управления паттернами битов"""
    print("\n=== Пример управления паттернами ===")
    
    plc = PLC210Client("http://localhost:5000")
    
    # Выключаем все биты
    print("1. Выключение всех битов...")
    if plc.turn_all(False):
        print("✅ Все биты выключены")
    else:
        print("❌ Ошибка выключения битов")
    
    time.sleep(1)
    
    # Включаем четные биты в регистре 0
    print("\n2. Включение четных битов в регистре 0...")
    for bit in range(0, 16, 2):  # 0, 2, 4, 6, 8, 10, 12, 14
        if plc.set_bit(0, bit, True):
            print(f"✅ Включен бит {bit}")
        else:
            print(f"❌ Ошибка включения бита {bit}")
        time.sleep(0.1)
    
    time.sleep(2)
    
    # Включаем нечетные биты в регистре 1
    print("\n3. Включение нечетных битов в регистре 1...")
    for bit in range(1, 16, 2):  # 1, 3, 5, 7, 9, 11, 13, 15
        if plc.set_bit(1, bit, True):
            print(f"✅ Включен бит {bit}")
        else:
            print(f"❌ Ошибка включения бита {bit}")
        time.sleep(0.1)

def example_register_operations():
    """Пример работы с регистрами целиком"""
    print("\n=== Пример работы с регистрами ===")
    
    plc = PLC210Client("http://localhost:5000")
    
    # Записываем различные паттерны в регистры
    patterns = {
        0: 0x5555,  # 0101010101010101
        1: 0xAAAA,  # 1010101010101010
        2: 0xFF00,  # 1111111100000000
        3: 0x00FF,  # 0000000011111111
        4: 0xF0F0,  # 1111000011110000
    }
    
    print("1. Запись паттернов в регистры...")
    for register, value in patterns.items():
        if plc.write_register(register, value):
            print(f"✅ Регистр {register}: {value} (0x{value:04X})")
        else:
            print(f"❌ Ошибка записи в регистр {register}")
        time.sleep(0.5)
    
    print("\n2. Чтение и проверка паттернов...")
    for register in patterns.keys():
        reg_info = plc.read_register(register)
        if reg_info and reg_info.get('success'):
            value = reg_info['value']
            expected = patterns[register]
            status = "✅" if value == expected else "❌"
            print(f"{status} Регистр {register}: {value} ({reg_info['hex']}) "
                  f"{'OK' if value == expected else f'Ожидалось {expected}'}")
        else:
            print(f"❌ Ошибка чтения регистра {register}")

def example_monitoring():
    """Пример мониторинга состояния"""
    print("\n=== Пример мониторинга ===")
    
    plc = PLC210Client("http://localhost:5000")
    
    print("Мониторинг состояния ПЛК210 (10 секунд)...")
    print("Нажмите Ctrl+C для остановки")
    
    try:
        start_time = time.time()
        while time.time() - start_time < 10:
            # Получаем полное состояние
            status = plc.get_status()
            if status and status.get('success'):
                data = status['data']
                timestamp = data['timestamp']
                connected = data['connection_status']
                
                print(f"\n[{timestamp}] Подключение: {'✅' if connected else '❌'}")
                
                # Показываем состояние каждого регистра
                for reg_name, reg_data in data['registers'].items():
                    if 'error' not in reg_data:
                        value = reg_data['value']
                        bits_on = sum(reg_data['bits'])
                        print(f"  {reg_name}: {value:5d} (0x{value:04X}) - битов ВКЛ: {bits_on:2d}")
                    else:
                        print(f"  {reg_name}: ОШИБКА - {reg_data['error']}")
            else:
                print("❌ Ошибка получения состояния")
            
            time.sleep(2)
            
    except KeyboardInterrupt:
        print("\n⏹️  Мониторинг остановлен")

def main():
    """Главная функция с примерами"""
    print("🏭 ПЛК210 - Примеры интеграции для Артема")
    print("=" * 50)
    
    try:
        # Базовое использование
        example_basic_usage()
        
        # Управление паттернами
        example_pattern_control()
        
        # Работа с регистрами
        example_register_operations()
        
        # Мониторинг
        example_monitoring()
        
    except KeyboardInterrupt:
        print("\n👋 Работа прервана пользователем")
    except Exception as e:
        print(f"\n❌ Ошибка: {e}")
    
    print("\n🎯 Примеры завершены!")
    print("📖 Подробная документация в README_ARTEM.md")

if __name__ == '__main__':
    main()
