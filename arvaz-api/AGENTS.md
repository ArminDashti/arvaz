## Learned User Preferences

- Keep the WebUI dark-theme only; do not reintroduce light mode or theme-switching code.
- Show "Dashti Technologies LLC" in the app chrome (sidebar / below logout).
- SoftEther Online auto-refresh is in minutes with choices 2 / 4 / 8 / 16 and default 4.
- Host realtime monitoring refresh choices are 1 / 2 / 4 seconds; keep polling light so it does not burn host resources.
- Prefer nav labels Mullvad (not VPN) and Docker (not Containers).
- Mullvad Quantum Resistance and DAITA are Mullvad tunnel features, not ISP logos.
- When using publish-on-t3 in this multi-root workspace, publish the project(s) that actually changed (often both).
- Ask before installing packages or other dependencies.

## Learned Workspace Facts

- This Cursor workspace is multi-root: `arvaz-api` (Go API) and `arvaz-webui` (Vue UI).
- Public hostnames in use include `arvaz.xaigrok.ir` (WebUI) and `arvaz-api.xaigrok.ir` (API).
- SoftEther runs in Docker beside the API; the API drives SoftEther via vpncmd against the SoftEther container (default hub `DEFAULT`).
- Host realtime metrics use gopsutil and target the Docker host (Ubuntu box), not only the API container cgroup view.
- Core WebUI areas include SoftEther Online / Users, Mullvad, Docker, Iperf (browser↔API LAN HTTP), and Host Monitor.
- Production updates commonly go through the publish-on-t3 / Docker-on-server path for one or both repos.
