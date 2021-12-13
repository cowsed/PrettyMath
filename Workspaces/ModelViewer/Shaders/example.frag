#version 330


in vec2 fragTexCoord;
in vec3 fragNormal;
in vec3 fragVert;
in vec3 fragWorldPos;

uniform vec3 lightpos;
uniform int ShadeNormal;
uniform float ambient;
uniform vec3 MaterialColor;

layout(location = 0) out vec4 frag_colour;
void main() {

    vec3 lpos=normalize(lightpos);
    float aspect = 8.0/6.0;
    vec3 col = vec3(1,0,0);
    float amt = dot(lpos, fragNormal);
    amt=ambient+(1-ambient)*amt;

    vec3 finalCol=vec3(1,1,1);
    if (ShadeNormal==1){
        finalCol = abs(fragNormal)*amt;
    } else if (ShadeNormal==2){
        finalCol=fract(fragWorldPos);
    } else if(ShadeNormal==3)
    {
        finalCol=MaterialColor*amt;
    } 
    frag_colour = vec4(finalCol, 1);
}