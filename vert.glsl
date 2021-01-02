#version 140
// in_Position was bound to attribute index 0 and in_Color was bound to attribute index 1

in vec3 in_Position;
in vec3 in_Color;
in vec3 in_Normal;
// We output the ex_Color variable to the next shader in the chain
out vec3 ex_Color;
out vec3 ex_Normal;
uniform vec3 lightPos;
void main(void) {
    // Since we are using flat lines, our input only had two points: x and y.
    // Set the Z coordinate to 0 and W coordinate to 1

    gl_Position = vec4(in_Position, 1.0);
    ex_Color = in_Color;
    // Normal = 
}