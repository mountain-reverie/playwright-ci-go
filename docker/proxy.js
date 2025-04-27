import * as net from 'net';

export function verify() {
    const client = new net.Socket();

    const host = 'host.testcontainers.internal';
    const port = process.argv[3];

    console.log('Connecting to server:', host, port);

    // Connect to the server
    client.connect(port, host, () => {
        console.log('Connected to server');

        // Send data to the server
        client.write('Hello from client!');
    });

    // Receive data from the server
    client.on('data', (data) => {
        console.log('Received from server:', data.toString());

        // Close the connection
        client.end();
    });

    // Handle errors
    client.on('error', (err) => {
        console.error('Connection error:', err);
        process.exit(1);
    });
}

