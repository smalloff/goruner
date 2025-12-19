export function GetConfig() {
  return window['go']['main']['App']['GetConfig']();
}

export function GetDiscoveredTests() {
  return window['go']['main']['App']['GetDiscoveredTests']();
}

export function RunTests() {
  return window['go']['main']['App']['RunTests']();
}

export function SaveConfig(arg1) {
  return window['go']['main']['App']['SaveConfig'](arg1);
}

export function TogglePause() {
  return window['go']['main']['App']['TogglePause']();
}