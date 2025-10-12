#!/usr/bin/env python3
"""
ПЛК210 Веб-сервер для управления Modbus TCP
Версия для разработчика Артема
Дата: 17 сентября 2025

Функциональность:
- REST API для управления битами в Holding Registers
- Веб-интерфейс для мониторинга и управления
- Поддержка 90 битов в 6 регистрах (0-5)
- Автоматическое переподключение при сбоях
"""

import socket
import struct
import logging
import json
import time
from datetime import datetime
from typing import List, Optional, Dict, Any
from dataclasses import dataclass
from flask import Flask, request, jsonify, render_template_string
from flask_cors import CORS
import threading

# Настройка логирования
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('plc210_webserver.log', encoding='utf-8'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

@dataclass
class ModbusConfig:
    """Конфигурация Modbus TCP подключения"""
    host: str = "195.208.131.189"
    port: int = 502
    unit_id: int = 1
    timeout: float = 5.0

class ModbusTCPClient:
    """Клиент для работы с Modbus TCP"""
    
    def __init__(self, config: ModbusConfig):
        self.config = config
        self.socket = None
        self.transaction_id = 1
        self.connected = False
        self.last_error = None
        self.connection_lock = threading.Lock()
        
    def connect(self) -> bool:
        """Подключение к ПЛК"""
        with self.connection_lock:
            try:
                if self.socket:
                    self.socket.close()
                
                self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                self.socket.settimeout(self.config.timeout)
                self.socket.connect((self.config.host, self.config.port))
                self.connected = True
                self.last_error = None
                logger.info(f"✓ Подключение к ПЛК210 установлено: {self.config.host}:{self.config.port}")
                return True
                
            except Exception as e:
                self.last_error = str(e)
                self.connected = False
                logger.error(f"✗ Ошибка подключения к ПЛК210: {e}")
                return False
    
    def disconnect(self):
        """Отключение от ПЛК"""
        with self.connection_lock:
            if self.socket:
                try:
                    self.socket.close()
                except:
                    pass
                finally:
                    self.socket = None
                    self.connected = False
                    logger.info("Отключение от ПЛК210")
    
    def _ensure_connection(self) -> bool:
        """Проверка и восстановление соединения"""
        if not self.connected:
            return self.connect()
        return True
    
    def _send_request(self, data: bytes) -> Optional[bytes]:
        """Отправка Modbus запроса"""
        if not self._ensure_connection():
            return None
            
        try:
            with self.connection_lock:
                self.socket.send(data)
                response = self.socket.recv(1024)
                return response
        except Exception as e:
            self.last_error = str(e)
            self.connected = False
            logger.error(f"Ошибка отправки запроса: {e}")
            return None
    
    def _create_modbus_header(self, pdu_length: int) -> bytes:
        """Создание заголовка Modbus TCP"""
        self.transaction_id = (self.transaction_id % 65535) + 1
        return struct.pack('>HHHB', 
                          self.transaction_id,  # Transaction ID
                          0,                    # Protocol ID
                          pdu_length + 1,       # Length
                          self.config.unit_id)  # Unit ID
    
    def read_holding_registers(self, address: int, count: int) -> Optional[List[int]]:
        """Чтение Holding Registers (Function 0x03)"""
        try:
            # PDU: Function Code (1) + Address (2) + Count (2) = 5 bytes
            header = self._create_modbus_header(5)
            pdu = struct.pack('>BHH', 0x03, address, count)  # Function 0x03
            request = header + pdu
            
            response = self._send_request(request)
            if not response or len(response) < 9:
                logger.error(f"Неполный ответ при чтении Holding Registers {address}")
                return None
            
            # Парсинг ответа: Header (7) + Function Code (1) + Byte Count (1) + Data
            byte_count = response[8]
            if len(response) < 9 + byte_count:
                logger.error(f"Недостаточно данных в ответе")
                return None
            
            values = []
            for i in range(count):
                offset = 9 + i * 2
                value = struct.unpack('>H', response[offset:offset+2])[0]
                values.append(value)
            
            logger.info(f"✓ Чтение Holding Registers {address}-{address+count-1}: {values}")
            return values
            
        except Exception as e:
            self.last_error = str(e)
            logger.error(f"Ошибка чтения Holding Registers {address}: {e}")
            return None
    
    def read_input_registers(self, address: int, count: int) -> Optional[List[int]]:
        """Чтение Input Registers (Function 0x04)"""
        try:
            # PDU: Function Code (1) + Address (2) + Count (2) = 5 bytes
            header = self._create_modbus_header(5)
            pdu = struct.pack('>BHH', 0x04, address, count)  # Function 0x04
            request = header + pdu
            
            response = self._send_request(request)
            if not response or len(response) < 9:
                logger.error(f"Неполный ответ при чтении Input Registers {address}")
                return None
            
            # Парсинг ответа: Header (7) + Function Code (1) + Byte Count (1) + Data
            byte_count = response[8]
            if len(response) < 9 + byte_count:
                logger.error(f"Недостаточно данных в ответе")
                return None
            
            values = []
            for i in range(count):
                offset = 9 + i * 2
                value = struct.unpack('>H', response[offset:offset+2])[0]
                values.append(value)
            
            logger.info(f"✓ Чтение Input Registers {address}-{address+count-1}: {values}")
            return values
            
        except Exception as e:
            self.last_error = str(e)
            logger.error(f"Ошибка чтения Input Registers {address}: {e}")
            return None
    
    def write_single_register(self, address: int, value: int) -> bool:
        """Запись в один Holding Register"""
        try:
            # PDU: Function Code (1) + Address (2) + Value (2) = 5 bytes
            header = self._create_modbus_header(5)
            pdu = struct.pack('>BHH', 0x06, address, value)  # Function 0x06
            request = header + pdu
            
            response = self._send_request(request)
            if not response:
                logger.error(f"Нет ответа при записи в Register {address}")
                return False
            
            # Для записи регистра ПЛК может отвечать не полностью
            # Считаем успехом если получили хоть что-то
            logger.info(f"✓ Запись в Holding Register {address} = {value}")
            return True
            
        except Exception as e:
            self.last_error = str(e)
            logger.error(f"Ошибка записи в Holding Register {address}: {e}")
            return False
    
    def get_bit(self, register: int, bit: int) -> Optional[bool]:
        """Получение состояния бита в регистре"""
        if not (0 <= register <= 5 and 0 <= bit <= 15):
            return None
            
        values = self.read_holding_registers(register, 1)
        if not values:
            return None
            
        return bool(values[0] & (1 << bit))
    
    def set_bit(self, register: int, bit: int, state: bool) -> bool:
        """Установка состояния бита в регистре"""
        if not (0 <= register <= 5 and 0 <= bit <= 15):
            return False
            
        # Читаем текущее значение
        values = self.read_holding_registers(register, 1)
        if not values:
            return False
        
        current_value = values[0]
        
        # Изменяем бит
        if state:
            new_value = current_value | (1 << bit)
        else:
            new_value = current_value & ~(1 << bit)
        
        # Записываем новое значение
        return self.write_single_register(register, new_value)
    
    def get_all_registers(self) -> Optional[Dict[str, Any]]:
        """Получение состояния всех регистров (Holding + Input)"""
        try:
            registers = {}
            
            for reg in range(6):
                # Чтение Holding Register (для команд)
                holding_values = self.read_holding_registers(reg, 1)
                
                # Чтение Input Register (для состояний)
                input_values = self.read_input_registers(reg, 1)
                
                # Определяем количество активных битов
                max_bits = 10 if reg == 5 else 16  # Register 5: только биты 0-9
                
                register_data = {
                    "holding_register": {
                        "value": holding_values[0] if holding_values else None,
                        "bits": [(holding_values[0] >> i) & 1 for i in range(16)] if holding_values else [None] * 16,
                        "available": holding_values is not None
                    },
                    "input_register": {
                        "value": input_values[0] if input_values else None,
                        "bits": [(input_values[0] >> i) & 1 for i in range(16)] if input_values else [None] * 16,
                        "available": input_values is not None
                    },
                    "max_bits": max_bits,
                    "overlay_working": (
                        holding_values is not None and 
                        input_values is not None and 
                        holding_values[0] == input_values[0]
                    ) if holding_values and input_values else False
                }
                
                # Для обратной совместимости - используем Input Register как основной источник состояния
                if input_values:
                    register_data["value"] = input_values[0]
                    register_data["bits"] = [(input_values[0] >> i) & 1 for i in range(16)]
                elif holding_values:
                    register_data["value"] = holding_values[0]
                    register_data["bits"] = [(holding_values[0] >> i) & 1 for i in range(16)]
                else:
                    register_data["value"] = None
                    register_data["bits"] = [None] * 16
                    register_data["error"] = "Не удалось прочитать ни Holding, ни Input Register"
                
                registers[f"register_{reg}"] = register_data
            
            return {
                "timestamp": datetime.now().isoformat(),
                "registers": registers,
                "connection_status": self.connected,
                "last_error": self.last_error
            }
        except Exception as e:
            logger.error(f"Ошибка получения состояния регистров: {e}")
            return None

# Создание Flask приложения
app = Flask(__name__)
CORS(app)

# Глобальный клиент Modbus
modbus_client = ModbusTCPClient(ModbusConfig())

# HTML шаблон для веб-интерфейса
HTML_TEMPLATE = """
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ПЛК210 Управление - Веб-интерфейс</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        
        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 15px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        
        .header {
            background: linear-gradient(135deg, #2c3e50 0%, #34495e 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        
        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
        }
        
        .status {
            display: flex;
            justify-content: center;
            align-items: center;
            gap: 15px;
            margin-top: 15px;
            font-size: 1.1em;
        }
        
        .status-indicator {
            width: 12px;
            height: 12px;
            border-radius: 50%;
            animation: pulse 2s infinite;
        }
        
        .status-connected {
            background: #2ecc71;
        }
        
        .status-disconnected {
            background: #e74c3c;
        }
        
        @keyframes pulse {
            0% { opacity: 1; }
            50% { opacity: 0.5; }
            100% { opacity: 1; }
        }
        
        .controls {
            padding: 30px;
        }
        
        .register-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
            gap: 25px;
            margin-bottom: 30px;
        }
        
        .register-card {
            background: #f8f9fa;
            border-radius: 12px;
            padding: 25px;
            border: 2px solid #e9ecef;
            transition: all 0.3s ease;
        }
        
        .register-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(0,0,0,0.1);
            border-color: #667eea;
        }
        
        .register-title {
            font-size: 1.3em;
            font-weight: bold;
            color: #2c3e50;
            margin-bottom: 15px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .register-value {
            background: #34495e;
            color: white;
            padding: 8px 12px;
            border-radius: 6px;
            font-family: 'Courier New', monospace;
            font-size: 0.9em;
        }
        
        .bits-grid {
            display: grid;
            grid-template-columns: repeat(8, 1fr);
            gap: 8px;
            margin-top: 15px;
        }
        
        .bit-button {
            aspect-ratio: 1;
            border: none;
            border-radius: 8px;
            font-weight: bold;
            font-size: 0.8em;
            cursor: pointer;
            transition: all 0.2s ease;
            position: relative;
            overflow: hidden;
        }
        
        .bit-button:hover {
            transform: scale(1.05);
        }
        
        .bit-button.bit-on {
            background: linear-gradient(135deg, #2ecc71 0%, #27ae60 100%);
            color: white;
            box-shadow: 0 4px 15px rgba(46, 204, 113, 0.4);
        }
        
        .bit-button.bit-off {
            background: linear-gradient(135deg, #ecf0f1 0%, #bdc3c7 100%);
            color: #2c3e50;
        }
        
        .bit-button:active {
            transform: scale(0.95);
        }
        
        .bit-button.overlay-ok {
            border: 2px solid #2ecc71;
        }
        
        .bit-button.overlay-warn {
            border: 2px solid #f39c12;
        }
        
        .actions {
            display: flex;
            justify-content: center;
            gap: 15px;
            flex-wrap: wrap;
            margin-top: 30px;
        }
        
        .btn {
            padding: 12px 24px;
            border: none;
            border-radius: 8px;
            font-size: 1em;
            font-weight: bold;
            cursor: pointer;
            transition: all 0.3s ease;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        
        .btn-primary {
            background: linear-gradient(135deg, #3498db 0%, #2980b9 100%);
            color: white;
        }
        
        .btn-success {
            background: linear-gradient(135deg, #2ecc71 0%, #27ae60 100%);
            color: white;
        }
        
        .btn-danger {
            background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%);
            color: white;
        }
        
        .btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(0,0,0,0.2);
        }
        
        .btn:active {
            transform: translateY(0);
        }
        
        .info-panel {
            background: #f8f9fa;
            border-radius: 12px;
            padding: 20px;
            margin-top: 25px;
            border-left: 5px solid #667eea;
        }
        
        .info-title {
            font-weight: bold;
            color: #2c3e50;
            margin-bottom: 10px;
        }
        
        .loading {
            display: none;
            text-align: center;
            padding: 20px;
            color: #667eea;
        }
        
        .spinner {
            display: inline-block;
            width: 20px;
            height: 20px;
            border: 3px solid #f3f3f3;
            border-top: 3px solid #667eea;
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin-right: 10px;
        }
        
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🏭 ПЛК210 Управление</h1>
            <p>Modbus TCP Веб-интерфейс</p>
            <div class="status">
                <div id="statusIndicator" class="status-indicator status-disconnected"></div>
                <span id="statusText">Подключение...</span>
                <span id="lastUpdate"></span>
            </div>
        </div>
        
        <div class="controls">
            <div class="loading" id="loading">
                <div class="spinner"></div>
                Загрузка данных...
            </div>
            
            <div id="registersContainer" class="register-grid">
                <!-- Регистры будут загружены динамически -->
            </div>
            
            <div class="actions">
                <button class="btn btn-primary" onclick="refreshData()">🔄 Обновить</button>
                <button class="btn btn-success" onclick="turnOnAll()">🟢 Все ВКЛ</button>
                <button class="btn btn-danger" onclick="turnOffAll()">🔴 Все ВЫКЛ</button>
            </div>
            
            <div class="info-panel">
                <div class="info-title">📋 Информация:</div>
                <p><strong>IP:</strong> 195.208.131.189:502</p>
                <p><strong>Unit ID:</strong> 1</p>
                <p><strong>Регистры:</strong> 0-5 (Holding Registers)</p>
                <p><strong>Биты:</strong> 90 бит управления (Register 0-5, биты 0-15, последние 6 битов Register 5 не используются)</p>
                <p><strong>Последнее обновление:</strong> <span id="timestampInfo">-</span></p>
            </div>
        </div>
    </div>

    <script>
        let currentData = null;
        
        // Автоматическое обновление каждые 2 секунды
        setInterval(refreshData, 2000);
        
        // Первоначальная загрузка
        refreshData();
        
        async function refreshData() {
            try {
                showLoading(true);
                const response = await fetch('/api/status');
                const data = await response.json();
                
                if (data.success) {
                    currentData = data.data;
                    updateInterface(data.data);
                    updateStatus(true, 'Подключено');
                } else {
                    updateStatus(false, 'Ошибка: ' + data.error);
                }
            } catch (error) {
                console.error('Ошибка получения данных:', error);
                updateStatus(false, 'Ошибка сети');
            } finally {
                showLoading(false);
            }
        }
        
        function updateInterface(data) {
            const container = document.getElementById('registersContainer');
            container.innerHTML = '';
            
            for (let reg = 0; reg < 6; reg++) {
                const registerData = data.registers[`register_${reg}`];
                const card = createRegisterCard(reg, registerData);
                container.appendChild(card);
            }
            
            // Обновляем timestamp
            document.getElementById('timestampInfo').textContent = 
                new Date(data.timestamp).toLocaleString('ru-RU');
        }
        
        function createRegisterCard(regNum, regData) {
            const card = document.createElement('div');
            card.className = 'register-card';
            
            const maxBits = regData.max_bits || (regNum === 5 ? 10 : 16);
            const overlayWorking = regData.overlay_working || false;
            const overlayStatus = overlayWorking ? '🔄 Overlay ✅' : '❌ Overlay';
            
            card.innerHTML = `
                <div class="register-title">
                    Register ${regNum} ${overlayStatus}
                    <div class="register-value">
                        ${regData.value !== null ? 
                            `0x${regData.value.toString(16).toUpperCase().padStart(4, '0')}` : 
                            'ERR'
                        }
                    </div>
                </div>
                ${regData.holding_register && regData.input_register ? `
                    <div style="font-size: 0.8em; margin-bottom: 10px; color: #666;">
                        H: ${regData.holding_register.available ? 
                            regData.holding_register.value : 'N/A'} | 
                        I: ${regData.input_register.available ? 
                            regData.input_register.value : 'N/A'}
                    </div>
                ` : ''}
                <div class="bits-grid">
                    ${Array.from({length: 16}, (_, bit) => {
                        if (bit >= maxBits) {
                            return `<div style="grid-column: span 1;"></div>`; // Пустое место
                        }
                        const isOn = regData.bits && regData.bits[bit];
                        const disabled = regData.error ? 'disabled' : '';
                        const bitClass = isOn ? 'bit-on' : 'bit-off';
                        const overlayClass = overlayWorking ? 'overlay-ok' : 'overlay-warn';
                        return `
                            <button class="bit-button ${bitClass} ${overlayClass}" 
                                    onclick="toggleBit(${regNum}, ${bit})" ${disabled}
                                    title="Bit ${bit}: ${isOn ? 'ВКЛ' : 'ВЫКЛ'} | Overlay: ${overlayWorking ? 'OK' : 'Warn'}">
                                ${bit}
                            </button>
                        `;
                    }).join('')}
                </div>
            `;
            
            return card;
        }
        
        async function toggleBit(register, bit) {
            try {
                if (!currentData || !currentData.registers[`register_${register}`]) {
                    alert('Нет данных о текущем состоянии регистра');
                    return;
                }
                
                const currentState = currentData.registers[`register_${register}`].bits[bit];
                const newState = !currentState;
                
                const response = await fetch('/api/set_bit', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        register: register,
                        bit: bit,
                        state: newState
                    })
                });
                
                const result = await response.json();
                
                if (result.success) {
                    // Обновляем интерфейс сразу
                    refreshData();
                } else {
                    alert('Ошибка: ' + result.error);
                }
            } catch (error) {
                console.error('Ошибка переключения бита:', error);
                alert('Ошибка сети');
            }
        }
        
        async function turnOnAll() {
            if (!confirm('Включить ВСЕ 90 битов?')) return;
            
            try {
                const response = await fetch('/api/turn_all', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({state: true})
                });
                
                const result = await response.json();
                
                if (result.success) {
                    refreshData();
                } else {
                    alert('Ошибка: ' + result.error);
                }
            } catch (error) {
                console.error('Ошибка включения всех битов:', error);
                alert('Ошибка сети');
            }
        }
        
        async function turnOffAll() {
            if (!confirm('ВЫКЛЮЧИТЬ все 90 битов?')) return;
            
            try {
                const response = await fetch('/api/turn_all', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({state: false})
                });
                
                const result = await response.json();
                
                if (result.success) {
                    refreshData();
                } else {
                    alert('Ошибка: ' + result.error);
                }
            } catch (error) {
                console.error('Ошибка выключения всех битов:', error);
                alert('Ошибка сети');
            }
        }
        
        function updateStatus(connected, message) {
            const indicator = document.getElementById('statusIndicator');
            const text = document.getElementById('statusText');
            const lastUpdate = document.getElementById('lastUpdate');
            
            indicator.className = `status-indicator ${connected ? 'status-connected' : 'status-disconnected'}`;
            text.textContent = message;
            lastUpdate.textContent = new Date().toLocaleTimeString('ru-RU');
        }
        
        function showLoading(show) {
            const loading = document.getElementById('loading');
            const container = document.getElementById('registersContainer');
            
            if (show) {
                loading.style.display = 'block';
                container.style.opacity = '0.5';
            } else {
                loading.style.display = 'none';
                container.style.opacity = '1';
            }
        }
    </script>
</body>
</html>
"""

# API Endpoints

@app.route('/')
def index():
    """Главная страница с веб-интерфейсом"""
    return render_template_string(HTML_TEMPLATE)

@app.route('/api/status')
def api_status():
    """Получение текущего состояния всех регистров"""
    try:
        data = modbus_client.get_all_registers()
        if data:
            return jsonify({
                "success": True,
                "data": data
            })
        else:
            return jsonify({
                "success": False,
                "error": "Не удалось получить данные с ПЛК"
            }), 500
    except Exception as e:
        logger.error(f"Ошибка API /status: {e}")
        return jsonify({
            "success": False,
            "error": str(e)
        }), 500

@app.route('/api/set_bit', methods=['POST'])
def api_set_bit():
    """Установка состояния конкретного бита"""
    try:
        data = request.get_json()
        
        register = data.get('register')
        bit = data.get('bit')
        state = data.get('state')
        
        if register is None or bit is None or state is None:
            return jsonify({
                "success": False,
                "error": "Отсутствуют обязательные параметры: register, bit, state"
            }), 400
        
        if not (0 <= register <= 5):
            return jsonify({
                "success": False,
                "error": "Номер регистра должен быть от 0 до 5"
            }), 400
        
        if not (0 <= bit <= 15):
            return jsonify({
                "success": False,
                "error": "Номер бита должен быть от 0 до 15"
            }), 400
        
        # Проверяем доступные биты для Register 5
        if register == 5 and bit >= 10:
            return jsonify({
                "success": False,
                "error": "В Register 5 доступны только биты 0-9"
            }), 400
        
        success = modbus_client.set_bit(register, bit, bool(state))
        
        if success:
            logger.info(f"✓ API: Установлен бит {bit} в Register {register} = {'ВКЛ' if state else 'ВЫКЛ'}")
            return jsonify({
                "success": True,
                "message": f"Бит {bit} в Register {register} установлен в {'ВКЛ' if state else 'ВЫКЛ'}"
            })
        else:
            return jsonify({
                "success": False,
                "error": "Не удалось установить бит"
            }), 500
            
    except Exception as e:
        logger.error(f"Ошибка API /set_bit: {e}")
        return jsonify({
            "success": False,
            "error": str(e)
        }), 500

@app.route('/api/get_bit')
def api_get_bit():
    """Получение состояния конкретного бита"""
    try:
        register = request.args.get('register', type=int)
        bit = request.args.get('bit', type=int)
        
        if register is None or bit is None:
            return jsonify({
                "success": False,
                "error": "Отсутствуют параметры: register, bit"
            }), 400
        
        if not (0 <= register <= 5 and 0 <= bit <= 15):
            return jsonify({
                "success": False,
                "error": "Неверные параметры register (0-5) или bit (0-15)"
            }), 400
        
        state = modbus_client.get_bit(register, bit)
        
        if state is not None:
            return jsonify({
                "success": True,
                "register": register,
                "bit": bit,
                "state": state
            })
        else:
            return jsonify({
                "success": False,
                "error": "Не удалось прочитать состояние бита"
            }), 500
            
    except Exception as e:
        logger.error(f"Ошибка API /get_bit: {e}")
        return jsonify({
            "success": False,
            "error": str(e)
        }), 500

@app.route('/api/turn_all', methods=['POST'])
def api_turn_all():
    """Включение/выключение всех 90 битов"""
    try:
        data = request.get_json()
        state = data.get('state')
        
        if state is None:
            return jsonify({
                "success": False,
                "error": "Отсутствует параметр state"
            }), 400
        
        success_count = 0
        total_count = 0
        
        # Регистры 0-4: все 16 битов
        for register in range(5):
            for bit in range(16):
                total_count += 1
                if modbus_client.set_bit(register, bit, bool(state)):
                    success_count += 1
        
        # Регистр 5: только биты 0-9 (10 битов)
        for bit in range(10):
            total_count += 1
            if modbus_client.set_bit(5, bit, bool(state)):
                success_count += 1
        
        action = "включены" if state else "выключены"
        
        if success_count == total_count:
            logger.info(f"✓ API: Все {total_count} битов {action}")
            return jsonify({
                "success": True,
                "message": f"Все {total_count} битов успешно {action}",
                "success_count": success_count,
                "total_count": total_count
            })
        else:
            logger.warning(f"⚠ API: {action} {success_count} из {total_count} битов")
            return jsonify({
                "success": False,
                "error": f"Удалось изменить только {success_count} из {total_count} битов",
                "success_count": success_count,
                "total_count": total_count
            }), 500
            
    except Exception as e:
        logger.error(f"Ошибка API /turn_all: {e}")
        return jsonify({
            "success": False,
            "error": str(e)
        }), 500

@app.route('/api/read_register')
def api_read_register():
    """Чтение значения регистра"""
    try:
        register = request.args.get('register', type=int)
        
        if register is None:
            return jsonify({
                "success": False,
                "error": "Отсутствует параметр register"
            }), 400
        
        if not (0 <= register <= 5):
            return jsonify({
                "success": False,
                "error": "Номер регистра должен быть от 0 до 5"
            }), 400
        
        values = modbus_client.read_holding_registers(register, 1)
        
        if values:
            value = values[0]
            return jsonify({
                "success": True,
                "register": register,
                "value": value,
                "hex": f"0x{value:04X}",
                "binary": f"{value:016b}",
                "bits": [(value >> i) & 1 for i in range(16)]
            })
        else:
            return jsonify({
                "success": False,
                "error": "Не удалось прочитать регистр"
            }), 500
            
    except Exception as e:
        logger.error(f"Ошибка API /read_register: {e}")
        return jsonify({
            "success": False,
            "error": str(e)
        }), 500

@app.route('/api/write_register', methods=['POST'])
def api_write_register():
    """Запись значения в регистр"""
    try:
        data = request.get_json()
        register = data.get('register')
        value = data.get('value')
        
        if register is None or value is None:
            return jsonify({
                "success": False,
                "error": "Отсутствуют параметры register, value"
            }), 400
        
        if not (0 <= register <= 5):
            return jsonify({
                "success": False,
                "error": "Номер регистра должен быть от 0 до 5"
            }), 400
        
        if not (0 <= value <= 65535):
            return jsonify({
                "success": False,
                "error": "Значение должно быть от 0 до 65535"
            }), 400
        
        success = modbus_client.write_single_register(register, value)
        
        if success:
            logger.info(f"✓ API: Записано значение {value} в Register {register}")
            return jsonify({
                "success": True,
                "message": f"Значение {value} записано в Register {register}",
                "register": register,
                "value": value,
                "hex": f"0x{value:04X}"
            })
        else:
            return jsonify({
                "success": False,
                "error": "Не удалось записать значение в регистр"
            }), 500
            
    except Exception as e:
        logger.error(f"Ошибка API /write_register: {e}")
        return jsonify({
            "success": False,
            "error": str(e)
        }), 500

@app.route('/api/read_input_register')
def api_read_input_register():
    """Чтение Input Register"""
    try:
        register = request.args.get('register', type=int)
        
        if register is None:
            return jsonify({
                "success": False,
                "error": "Отсутствует параметр register"
            }), 400
        
        if not (0 <= register <= 5):
            return jsonify({
                "success": False,
                "error": "Номер регистра должен быть от 0 до 5"
            }), 400
        
        values = modbus_client.read_input_registers(register, 1)
        
        if values:
            value = values[0]
            max_bits = 10 if register == 5 else 16
            return jsonify({
                "success": True,
                "register": register,
                "value": value,
                "hex": f"0x{value:04X}",
                "binary": f"{value:016b}",
                "bits": [(value >> i) & 1 for i in range(16)],
                "max_bits": max_bits,
                "register_type": "input"
            })
        else:
            return jsonify({
                "success": False,
                "error": "Не удалось прочитать Input Register"
            }), 500
            
    except Exception as e:
        logger.error(f"Ошибка API /read_input_register: {e}")
        return jsonify({
            "success": False,
            "error": str(e)
        }), 500

@app.route('/api/overlay_status')
def api_overlay_status():
    """Проверка статуса overlay Holding + Input Registers"""
    try:
        overlay_results = {}
        
        for reg in range(6):
            holding_values = modbus_client.read_holding_registers(reg, 1)
            input_values = modbus_client.read_input_registers(reg, 1)
            
            overlay_results[f"register_{reg}"] = {
                "holding_available": holding_values is not None,
                "input_available": input_values is not None,
                "holding_value": holding_values[0] if holding_values else None,
                "input_value": input_values[0] if input_values else None,
                "overlay_working": (
                    holding_values is not None and 
                    input_values is not None and 
                    holding_values[0] == input_values[0]
                ),
                "max_bits": 10 if reg == 5 else 16
            }
        
        total_overlay_working = sum(
            1 for result in overlay_results.values() 
            if result["overlay_working"]
        )
        
        return jsonify({
            "success": True,
            "overlay_status": {
                "total_registers": 6,
                "overlay_working_count": total_overlay_working,
                "overlay_working_percentage": (total_overlay_working / 6) * 100,
                "all_overlay_working": total_overlay_working == 6
            },
            "registers": overlay_results
        })
        
    except Exception as e:
        logger.error(f"Ошибка API /overlay_status: {e}")
        return jsonify({
            "success": False,
            "error": str(e)
        }), 500

@app.route('/api/connection')
def api_connection():
    """Информация о подключении"""
    return jsonify({
        "success": True,
        "connection": {
            "host": modbus_client.config.host,
            "port": modbus_client.config.port,
            "unit_id": modbus_client.config.unit_id,
            "timeout": modbus_client.config.timeout,
            "connected": modbus_client.connected,
            "last_error": modbus_client.last_error
        }
    })

def main():
    """Запуск веб-сервера"""
    logger.info("🚀 Запуск ПЛК210 Веб-сервера...")
    
    # Попытка подключения к ПЛК при старте
    if modbus_client.connect():
        logger.info("✅ Начальное подключение к ПЛК210 успешно")
    else:
        logger.warning("⚠️ Не удалось подключиться к ПЛК210 при старте (будет переподключение)")
    
    try:
        # Запуск Flask сервера
        logger.info("🌐 Веб-сервер запускается на http://0.0.0.0:5001")
        app.run(
            host='0.0.0.0',
            port=5001,
            debug=False,
            threaded=True
        )
    except KeyboardInterrupt:
        logger.info("🛑 Остановка сервера по запросу пользователя")
    except Exception as e:
        logger.error(f"❌ Ошибка запуска сервера: {e}")
    finally:
        modbus_client.disconnect()
        logger.info("👋 Веб-сервер остановлен")

if __name__ == '__main__':
    main()
