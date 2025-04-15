document.addEventListener('DOMContentLoaded', function() {
    // Source selection
    document.querySelectorAll('input[name="source"]').forEach(radio => {
        radio.addEventListener('change', function() {
            document.getElementById('clickhouse-config').style.display = 
                this.value === 'clickhouse' ? 'block' : 'none';
            document.getElementById('file-config').style.display = 
                this.value === 'file' ? 'block' : 'none';
        });
    });

    // ClickHouse connection
    document.getElementById('ch-connect').addEventListener('click', async function() {
        const host = document.getElementById('ch-host').value;
        const port = document.getElementById('ch-port').value;
        const database = document.getElementById('ch-database').value;
        const user = document.getElementById('ch-user').value;
        const token = document.getElementById('ch-token').value;
        const secure = document.getElementById('ch-secure').checked;

        updateStatus('Connecting to ClickHouse...');
        
        try {
            const response = await fetch('/api/connect', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    host, port, database, user, jwtToken: token, secure
                })
            });

            const data = await response.json();
            if (response.ok) {
                updateStatus('Connected successfully! Discovering schema...');
                await discoverSchema();
            } else {
                updateStatus(`Error: ${data.error || 'Connection failed'}`);
            }
        } catch (error) {
            updateStatus(`Error: ${error.message}`);
        }
    });

    // File loading
    document.getElementById('file-load').addEventListener('click', async function() {
        const filePath = document.getElementById('file-path').value;
        const delimiter = document.getElementById('file-delimiter').value;
        
        updateStatus('Loading file schema...');
        
        try {
            const response = await fetch('/api/discover-schema', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    filePath,
                    delimiter
                })
            });

            const data = await response.json();
            if (response.ok) {
                updateStatus('File loaded successfully!');
                displayColumns(data.columns);
            } else {
                updateStatus(`Error: ${data.error || 'File loading failed'}`);
            }
        } catch (error) {
            updateStatus(`Error: ${error.message}`);
        }
    });

    // Start transfer
    document.getElementById('start-transfer').addEventListener('click', async function() {
        const selectedColumns = Array.from(
            document.querySelectorAll('#columns-list input[type="checkbox"]:checked')
        ).map(el => el.value);

        updateStatus('Starting data transfer...');
        
        try {
            const response = await fetch('/api/transfer', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    columns: selectedColumns
                })
            });

            const data = await response.json();
            if (response.ok) {
                updateStatus('Transfer completed successfully!');
                updateResult(`Records processed: ${data.recordCount}`);
            } else {
                updateStatus(`Error: ${data.error || 'Transfer failed'}`);
            }
        } catch (error) {
            updateStatus(`Error: ${error.message}`);
        }
    });

    async function discoverSchema() {
        try {
            const response = await fetch('/api/discover-schema');
            const data = await response.json();
            
            if (response.ok) {
                displayTables(data.tables);
            } else {
                updateStatus(`Error: ${data.error || 'Schema discovery failed'}`);
            }
        } catch (error) {
            updateStatus(`Error: ${error.message}`);
        }
    }

    function displayTables(tables) {
        const tablesList = document.getElementById('tables-list');
        tablesList.innerHTML = '<h3>Tables</h3>';
        
        tables.forEach(table => {
            const div = document.createElement('div');
            div.innerHTML = `
                <label>
                    <input type="radio" name="table" value="${table}">
                    ${table}
                </label>
            `;
            div.querySelector('input').addEventListener('change', async function() {
                if (this.checked) {
                    await loadTableColumns(table);
                }
            });
            tablesList.appendChild(div);
        });
        
        document.getElementById('schema-section').style.display = 'block';
    }

    async function loadTableColumns(table) {
        updateStatus(`Loading columns for ${table}...`);
        
        try {
            const response = await fetch(`/api/discover-schema?table=${table}`);
            const data = await response.json();
            
            if (response.ok) {
                displayColumns(data.columns);
            } else {
                updateStatus(`Error: ${data.error || 'Column loading failed'}`);
            }
        } catch (error) {
            updateStatus(`Error: ${error.message}`);
        }
    }

    function displayColumns(columns) {
        const columnsList = document.getElementById('columns-list');
        columnsList.innerHTML = '<h3>Columns</h3>';
        
        columns.forEach(column => {
            const div = document.createElement('div');
            div.innerHTML = `
                <label>
                    <input type="checkbox" value="${column}" checked>
                    ${column}
                </label>
            `;
            columnsList.appendChild(div);
        });
    }

    function updateStatus(message) {
        document.getElementById('status-message').textContent = message;
    }

    function updateResult(message) {
        document.getElementById('result-message').textContent = message;
    }
});