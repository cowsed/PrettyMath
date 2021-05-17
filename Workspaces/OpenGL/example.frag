#version 330
in vec2 UV;

layout(location = 0) out vec4 frag_colour;
uniform float rotation; 
uniform  sampler2D tex;
void main() {
    float aspect = 8.0/6.0;
    vec2 uv = UV/vec2(aspect, 1);
    float r,f,a;
    uv*=2;
    r=length(uv);
    a = atan(uv.y,uv.x)+rotation;
    f = sin(a*6);
    vec3 col = vec3(step(f,r));
    col*=texture(tex, uv.xy).rgb;
    frag_colour = vec4(col, 1);
}