import React, { useState } from 'react';
import CodeBlock from '@theme/CodeBlock';

const ClickHouseDSNGenerator = () => {
    const [formData, setFormData] = useState({
        hostname: '<hostname>',
        port: '<port>',
        username: 'default',
        password: '<password>',
        isCloud: true,
        useHttps: true
    });

    const [copied, setCopied] = useState(false);
    const [outputFormat, setOutputFormat] = useState('yaml'); // 'yaml' | 'env'

    const handleInputChange = (e) => {
        const { name, value, type, checked } = e.target;
        setFormData(prev => {
            const next = {
                ...prev,
                [name]: type === 'checkbox' ? checked : value,
            };
            // If HTTPS is turned off, also turn off ClickHouse Cloud
            if (name === 'useHttps' && !checked) {
                next.isCloud = false;
            }
            return next;
        });
    };

    const generateDSNUrl = () => {
        const { hostname, port, username, password, isCloud, useHttps } = formData;

        if (!hostname || !username || !password) {
            return '';
        }

        let dsn;

        if (useHttps) {
            // HTTPS format for ClickHouse Cloud
            const baseUrl = hostname.startsWith('http') ? hostname : `https://${hostname}`;
            const params = new URLSearchParams({
                username,
                password,
                ...(isCloud && { secure: 'true', skip_verify: 'true' })
            });
            dsn = `${baseUrl}:${port}?${params.toString()}`;
        } else {
            // Standard clickhouse:// format
            const params = new URLSearchParams({
                username,
                password
            });
            dsn = `clickhouse://${hostname}:${port}?${params.toString()}`;
        }

        return dsn;
    };

    const dsnUrl = generateDSNUrl();
    const snippet = dsnUrl
        ? (outputFormat === 'yaml'
            ? `${dsnUrl}`
            : `connector.clickhouse.dsn="${dsnUrl}"`)
        : '';

    const copyToClipboard = async () => {
        if (!snippet) return;

        try {
            await navigator.clipboard.writeText(snippet);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch (err) {
            console.error('Failed to copy: ', err);
        }
    };

    return (
        <div className="ch_container">
            <div className="formSection">
                {/* <h4>Generate ClickHouse DSN</h4> */}

                <div className="formGrid">
                    <div className="inputGroup">
                        <label htmlFor="hostname">Hostname</label>
                        <input
                            type="text"
                            id="hostname"
                            name="hostname"
                            value={formData.hostname}
                            onChange={handleInputChange}
                            placeholder="unique-id.region.aws.clickhouse.cloud"
                        />
                    </div>

                    <div className="inputGroup">
                        <label htmlFor="port">Port</label>
                        <input
                            type="text"
                            id="port"
                            name="port"
                            value={formData.port}
                            onChange={handleInputChange}
                            placeholder="8443"
                        />
                    </div>

                    <div className="inputGroup">
                        <label htmlFor="username">Username</label>
                        <input
                            type="text"
                            id="username"
                            name="username"
                            value={formData.username}
                            onChange={handleInputChange}
                            placeholder="default"
                        />
                    </div>

                    <div className="inputGroup">
                        <label htmlFor="password">Password</label>
                        <input
                            type="password"
                            id="password"
                            name="password"
                            value={formData.password}
                            onChange={handleInputChange}
                            placeholder="your_password"
                        />
                    </div>
                </div>

                <div className="checkboxGroup">
                    <label className="checkbox">
                        <input
                            type="checkbox"
                            name="useHttps"
                            checked={formData.useHttps}
                            onChange={handleInputChange}
                        />
                        Use HTTPS (recommended for ClickHouse Cloud)
                    </label>

                    <label className="checkbox">
                        <input
                            type="checkbox"
                            name="isCloud"
                            checked={formData.isCloud}
                            onChange={handleInputChange}
                            disabled={!formData.useHttps}
                            title={!formData.useHttps ? 'Enable HTTPS to use ClickHouse Cloud' : undefined}
                        />
                        ClickHouse Cloud (adds secure=true&skip_verify=true)
                    </label>
                </div>

                <div className="checkboxGroup">
                    {/* <span className="outputHeaderLabel">Output format</span> */}
                    <div className="segmentedControl" role="tablist" aria-label="Output format">
                        <div className="segmentOption">
                            <input
                                className="visuallyHidden"
                                type="radio"
                                id="format-yaml"
                                name="outputFormat"
                                value="yaml"
                                checked={outputFormat === 'yaml'}
                                onChange={() => setOutputFormat('yaml')}
                            />
                            <label
                                htmlFor="format-yaml"
                                className={`segmentLabel ${outputFormat === 'yaml' ? 'selected' : ''}`}
                                role="tab"
                                aria-selected={outputFormat === 'yaml'}
                            >
                                Connection String
                            </label>
                        </div>
                        <div className="segmentOption">
                            <input
                                className="visuallyHidden"
                                type="radio"
                                id="format-env"
                                name="outputFormat"
                                value="env"
                                checked={outputFormat === 'env'}
                                onChange={() => setOutputFormat('env')}
                            />
                            <label
                                htmlFor="format-env"
                                className={`segmentLabel ${outputFormat === 'env' ? 'selected' : ''}`}
                                role="tab"
                                aria-selected={outputFormat === 'env'}
                            >
                                .env
                            </label>
                        </div>
                    </div>
                </div>
            </div>

            <div className="outputSection">
                <div className="outputHeader">
                </div>
                <CodeBlock language={outputFormat === 'yaml' ? 'yaml' : 'bash'}>{snippet}</CodeBlock>
            </div>
        </div>
    );
};

export default ClickHouseDSNGenerator; 