#version 330


in vec2 fragTexCoord;
in vec3 fragNormal;
in vec3 fragVert;
in vec3 fragWorldPos;

uniform vec3 lightpos;
uniform vec3 viewpos;
uniform int ShadeNormal;
uniform float ambient;
uniform vec3 MaterialColor;
vec3 lightColor = vec3(1,1,1);

layout(location = 0) out vec4 frag_colour;
void main() {
    vec3 lightDir   = normalize(lightpos - fragWorldPos);
    vec3 viewDir    = normalize(viewpos - fragWorldPos);
    vec3 halfwayDir = normalize(lightDir + viewDir);


    vec3 lpos=normalize(lightpos);
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

