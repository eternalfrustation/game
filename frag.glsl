#version 140
// It was expressed that some drivers required this next line to function properly
precision highp float;

in vec3 ex_Color;
in vec3 ex_Normal;
void main(void) {
    // Pass through our original color with full opacity.
    float ambientStrength = 1;
    vec3 lightColor = vec3(0.85,0.8,0.9);
    vec3 ambient = ambientStrength*lightColor;
    gl_FragColor = vec4(ex_Color*ambient,1.0);
    // gl_FrontColor = vec4(ex_Color, 1.0);
}