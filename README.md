<p align="center">
  <img src="images/tailpass.png" alt="Tailpass Logo" width="450">
</p>

# PortPass
Tailpass is a Tailscale powered TCP port forwarding tool that bridges your VLANs, containers, and hosts  simply and securely.
You can easily connect web servers, SSH sessions, databases, or any TCP service across your network without worrying about complex configurations.
Add your local and remote services, start the tunnel, and your traffic flows seamlessly through Tailscale. Tailpass gives you a lightweight dashboard, an efficient backend, and the freedom to access your services from anywhere.

## Architecture

<p align="center">
  <img src="images/Architecture.svg" alt="architecture" width="600">
</p>

## Screenshots

### Dashboard

![Dashboard](images/Dashboard.png)

### Create

![Create](images/Create.png)


### Service access

![SSH_Access](images/SSH_Access.png)


### Access TailPass with Phone

<p align="left">
  <img src="images/TailPass_with_termux.jpg" alt="Tailpass" width="300">
</p>


<p align="left">
  <img src="images/Tailscale_Phone.jpg" alt="Tailpass" width="300">
</p>

<p align="left">
  <img src="images/Dashboard_on_Phone.jpg" alt="Tailpass" width="300">
</p>

## Roadmap

Fix and Features
---

### Planned Enhancements
- [ ] **GUI Improvements**
  - Refine layout and responsiveness  
  - Improve service management UX  

- [ ] **Health Checks**
  - Implement port-based health monitoring for each service  
  - Periodic background checks with automatic status updates  

- [ ] **Authentication**
  - Add **Login Page** with **JWT-based authentication**  
  - Secure routes for authenticated users only  
  - Token refresh and logout feature  

---

###  Bug Fixes   
- [ ] Clean up unused CSS/JS dependencies  
- [ ] Handle missing or invalid environment variables gracefully  
- [ ] Optimize database connection pool size and error handling  

---

### Future Features
- [ ] **Service Restart & Stop Controls** (from GUI)  
- [ ] **Container Health Metrics Dashboard** (CPU, memory, uptime)  
- [ ] **Configurable Alerts** (email/webhook on service failure)  
---





