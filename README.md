# Gotify MQTT Forwarding Plugin

![Version Badge](https://img.shields.io/badge/Version-0.0.1-blue)
![License Badge](https://img.shields.io/badge/License-MIT-green)

A Gotify plugin for MQTT message forwarding with customizable settings.

## Description

This plugin allows you to seamlessly integrate Gotify with MQTT. It provides the ability to forward messages received on Gotify to an MQTT broker, with options to customize various MQTT settings.

## Features

- Send Gotify messages to an MQTT broker.
- Configurable MQTT settings, including:
  - IP and Port of the broker
  - Quality of Service (QoS)
  - Topic for publishing
  - KeepAlive duration
  - Message retain option

## Installation

1. Clone this repository.
2. Build the plugin using Go.
3. Move the compiled plugin to Gotify's plugin directory.
4. Restart Gotify.

## Configuration

After installing the plugin, you can configure it through Gotify's web interface. Navigate to the Plugins section, and you should see the `gotify/mqtt` plugin listed. Click on it to configure the MQTT settings as per your broker's setup.

## Contributing

If you find any bugs or would like to see new features, please open an issue or submit a pull request.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Links

- [Gotify's Official Website](https://gotify.net/)
- [MQTT Protocol Official Website](http://mqtt.org/)

## Author

[FaintGhost](https://github.com/FaintGhost)
